package activeprobe

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"s3intel/internal/classifier"
	"s3intel/internal/secretscan"
)

// S3API, gerçek aws-sdk-go-v2 S3 istemcisinin bu pakette kullanılan
// alt kümesidir. Test edilebilirlik için bir arayüz olarak tanımlanır:
// probe_test.go bu arayüzü gerçek AWS'ye hiç dokunmadan bir mock ile
// doldurup EnumerateBucket'ın whitelist reddini kanıtlayabilir.
type S3API interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// ObjectFinding, bir S3 nesnesi için toplanan sınıflandırma + sır tarama
// sonucudur. riskengine paketi bunu nihai risk skoruna çevirir.
type ObjectFinding struct {
	Key            string
	Size           int64
	Category       classifier.Category
	BaseRisk       int
	SecretFindings []secretscan.Finding
}

// maxScannableObjectSize, içerik taramasının yapılacağı üst boyut sınırıdır.
// Bunun üzerindeki (ör. video/imaj) nesneler sadece sınıflandırılır, metin
// olarak indirilip taranmaz.
const maxScannableObjectSize = 5 * 1024 * 1024 // 5 MB

func newRealS3Client(ctx context.Context) (S3API, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("AWS yapılandırması yüklenemedi: %w", err)
	}
	return s3.NewFromConfig(cfg), nil
}

// s3ClientFactory, gerçek S3 istemcisini üreten fonksiyondur. Testler bu
// paket değişkenini bir mock fabrikasıyla değiştirip EnumerateBucket'ın
// whitelist reddedildiğinde bu fonksiyonu HİÇ çağırmadığını doğrular.
var s3ClientFactory = newRealS3Client

// EnumerateBucket, whitelist'teki bir bucket'ı listeler, her nesnenin
// içeriğini indirir, sınıflandırır ve içinde sır arar.
//
// KRİTİK MİMARİ KURAL: Whitelist kontrolü bu fonksiyonun İLK SATIRIDIR.
// Kontrol ile S3 istemcisinin oluşturulması/çağrılması arasında BAŞKA HİÇBİR
// İŞ MANTIĞI YOKTUR — whitelist reddederse s3ClientFactory hiçbir zaman
// çağrılmaz, dolayısıyla hiçbir AWS SDK ağ çağrısı asla tetiklenmez.
func EnumerateBucket(ctx context.Context, bucketName string) ([]ObjectFinding, error) {
	if !IsAllowed(bucketName) {
		return nil, fmt.Errorf("bucket %s is not in the whitelist — active probing refused", bucketName)
	}
	client, err := s3ClientFactory(ctx)
	if err != nil {
		return nil, err
	}
	return enumerateWithClient(ctx, client, bucketName)
}

func enumerateWithClient(ctx context.Context, client S3API, bucketName string) ([]ObjectFinding, error) {
	var findings []ObjectFinding
	var continuationToken *string

	for {
		out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucketName),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("bucket %s listelenemedi: %w", bucketName, err)
		}

		for _, obj := range out.Contents {
			key := aws.ToString(obj.Key)
			category, baseRisk := classifier.Classify(key)

			var secrets []secretscan.Finding
			size := aws.ToInt64(obj.Size)
			if size > 0 && size <= maxScannableObjectSize {
				content, err := getObjectContent(ctx, client, bucketName, key)
				if err != nil {
					return nil, fmt.Errorf("nesne indirilemedi %s/%s: %w", bucketName, key, err)
				}
				secrets = secretscan.Scan(content)
			}

			findings = append(findings, ObjectFinding{
				Key:            key,
				Size:           size,
				Category:       category,
				BaseRisk:       baseRisk,
				SecretFindings: secrets,
			})
		}

		if out.IsTruncated == nil || !*out.IsTruncated {
			break
		}
		continuationToken = out.NextContinuationToken
	}

	return findings, nil
}

func getObjectContent(ctx context.Context, client S3API, bucketName, key string) (string, error) {
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
