package activeprobe

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// mockS3Client, gerçek AWS'ye hiç dokunmayan sahte bir S3API uygulamasıdır.
// listCalled/getCalled bayrakları, çağrıların gerçekten tetiklenip
// tetiklenmediğini kanıtlamak için kullanılır.
type mockS3Client struct {
	listCalled bool
	getCalled  bool
	listOutput *s3.ListObjectsV2Output
	getBodies  map[string]string
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	m.listCalled = true
	return m.listOutput, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.getCalled = true
	body := m.getBodies[aws.ToString(params.Key)]
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(body)),
	}, nil
}

func withMockFactory(t *testing.T, client S3API) {
	t.Helper()
	original := s3ClientFactory
	s3ClientFactory = func(ctx context.Context) (S3API, error) {
		return client, nil
	}
	t.Cleanup(func() { s3ClientFactory = original })
}

// TestEnumerateBucket_RejectsNonWhitelistedBucket, whitelist'te olmayan bir
// bucket adıyla EnumerateBucket çağrıldığında GERÇEK (ya da mock) hiçbir S3
// çağrısının tetiklenmediğini kanıtlar. Bu, PROBE.GO'DAKİ WHITELIST
// KONTROLÜNÜN AWS ÇAĞRISINDAN ÖNCE GELDİĞİNİN unit test kanıtıdır.
func TestEnumerateBucket_RejectsNonWhitelistedBucket(t *testing.T) {
	SetWhitelistPath("testdata/whitelist_test.yaml")
	t.Cleanup(func() { SetWhitelistPath("") })

	mock := &mockS3Client{}
	withMockFactory(t, mock)

	_, err := EnumerateBucket(context.Background(), "not-whitelisted-bucket")
	if err == nil {
		t.Fatal("whitelist'te olmayan bucket için hata bekleniyordu, nil döndü")
	}
	if !strings.Contains(err.Error(), "whitelist") {
		t.Fatalf("hata mesajı whitelist reddini belirtmeli, aldık: %v", err)
	}
	if mock.listCalled {
		t.Error("whitelist reddedilince ListObjectsV2 HİÇ çağrılmamalıydı, ama çağrıldı")
	}
	if mock.getCalled {
		t.Error("whitelist reddedilince GetObject HİÇ çağrılmamalıydı, ama çağrıldı")
	}
}

// TestEnumerateBucket_AllowsWhitelistedBucket, whitelist'teki bir bucket için
// akışın gerçekten S3 istemcisini çağırdığını ve içerik taramasının
// (secretscan) devreye girdiğini doğrular — hepsi mock istemci üzerinden,
// gerçek ağ çağrısı olmadan.
func TestEnumerateBucket_AllowsWhitelistedBucket(t *testing.T) {
	SetWhitelistPath("testdata/whitelist_test.yaml")
	t.Cleanup(func() { SetWhitelistPath("") })

	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []types.Object{
				{Key: aws.String("config/.env"), Size: aws.Int64(42)},
			},
		},
		getBodies: map[string]string{
			"config/.env": "AWS_ACCESS_KEY_ID=AKIAABCDEFGHIJKLMNOP\n",
		},
	}
	withMockFactory(t, mock)

	findings, err := EnumerateBucket(context.Background(), "test-lab-level1")
	if err != nil {
		t.Fatalf("beklenmeyen hata: %v", err)
	}
	if !mock.listCalled {
		t.Error("whitelist'teki bucket için ListObjectsV2 çağrılmalıydı")
	}
	if !mock.getCalled {
		t.Error("küçük boyutlu nesne için GetObject çağrılmalıydı")
	}
	if len(findings) != 1 {
		t.Fatalf("1 bulgu bekleniyordu, %d geldi", len(findings))
	}
	if len(findings[0].SecretFindings) == 0 {
		t.Error("AKIA deseni içeren dosyada secret bulgusu bekleniyordu")
	}
}
