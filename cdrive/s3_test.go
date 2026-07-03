package drive

import (
	"context"
	"testing"
)

func TestS3ConfigUsesLegacyCloudFrontBaseURL(t *testing.T) {
	drive := initS3Drive(S3Config{
		Region:            "eu-south-1",
		BucketName:        "uc-core",
		CloudFrontBaseURL: "https://cdn.univus.cloud/core",
	})

	if got, want := drive.objectKeyFullPath("avatars/a.png"), "core/avatars/a.png"; got != want {
		t.Fatalf("objectKeyFullPath() = %q, want %q", got, want)
	}
	if got, want := drive.objectKeyFullPath("core/avatars/a.png"), "core/avatars/a.png"; got != want {
		t.Fatalf("objectKeyFullPath() with prefixed key = %q, want %q", got, want)
	}
	if got, want := drive.publicURL("core/avatars/a.png"), "https://cdn.univus.cloud/core/avatars/a.png"; got != want {
		t.Fatalf("publicURL() = %q, want %q", got, want)
	}
	if got, want := drive.config.provider(), string(S3DriveOption); got != want {
		t.Fatalf("provider() = %q, want %q", got, want)
	}
}

func TestS3ConfigSupportsR2EndpointAndPublicBaseURL(t *testing.T) {
	drive := initS3Drive(S3Config{
		BucketName:    "uc-core",
		Endpoint:      "https://account-id.r2.cloudflarestorage.com",
		PublicBaseURL: "https://cdn.univus.cloud/core/",
		KeyPrefix:     "/core/",
		AccessKeyID:   "access-key",
	})

	if got, want := drive.config.region(), "auto"; got != want {
		t.Fatalf("region() = %q, want %q", got, want)
	}
	if got, want := drive.config.provider(), string(R2DriveOption); got != want {
		t.Fatalf("provider() = %q, want %q", got, want)
	}
	if got, want := drive.objectKeyFullPath("avatars/a.png"), "core/avatars/a.png"; got != want {
		t.Fatalf("objectKeyFullPath() = %q, want %q", got, want)
	}
	if got, want := drive.objectKeyFullPath("core/avatars/a.png"), "core/avatars/a.png"; got != want {
		t.Fatalf("objectKeyFullPath() with prefixed key = %q, want %q", got, want)
	}
	if got, want := drive.publicURL("core/avatars/a.png"), "https://cdn.univus.cloud/core/avatars/a.png"; got != want {
		t.Fatalf("publicURL() = %q, want %q", got, want)
	}
}

func TestS3ConfigAllowsKeyPrefixWithoutPublicURLPath(t *testing.T) {
	drive := initS3Drive(S3Config{
		Endpoint:      "https://account-id.r2.cloudflarestorage.com",
		PublicBaseURL: "https://cdn.univus.cloud",
		KeyPrefix:     "core",
	})

	if got, want := drive.objectKeyFullPath("avatars/a.png"), "core/avatars/a.png"; got != want {
		t.Fatalf("objectKeyFullPath() = %q, want %q", got, want)
	}
	if got, want := drive.publicURL("core/avatars/a.png"), "https://cdn.univus.cloud/core/avatars/a.png"; got != want {
		t.Fatalf("publicURL() = %q, want %q", got, want)
	}
}

func TestBaseDriveS3UsesRequestScopedContext(t *testing.T) {
	base := InitWithConfig(Config{S3: S3Config{Region: "eu-south-1"}})
	ctx := context.WithValue(context.Background(), "request-id", "one")

	got := base.S3(ctx)
	backend, ok := got.backend.(*S3Drive)
	if !ok {
		t.Fatalf("S3() backend = %T, want *S3Drive", got.backend)
	}

	if backend == base.s3 {
		t.Fatalf("S3() returned shared backend instance")
	}
	if backend.ctx != ctx {
		t.Fatalf("S3() did not bind request context")
	}
	if base.s3.ctx == ctx {
		t.Fatalf("S3() mutated shared drive context")
	}
}

func TestBaseDriveObjectFacadeSelectsBackends(t *testing.T) {
	base := InitWithConfig(Config{S3: S3Config{Region: "eu-south-1"}})
	ctx := context.Background()

	if got := base.Object(ctx, FSDriveOption); got.Option() != FSDriveOption {
		t.Fatalf("Object(fs).Option() = %q, want %q", got.Option(), FSDriveOption)
	}
	if _, ok := base.Object(ctx, FSDriveOption).backend.(*FSDrive); !ok {
		t.Fatalf("Object(fs) backend = %T, want *FSDrive", base.Object(ctx, FSDriveOption).backend)
	}
	if got := base.Object(ctx, S3DriveOption); got.Option() != S3DriveOption {
		t.Fatalf("Object(s3).Option() = %q, want %q", got.Option(), S3DriveOption)
	}
	if _, ok := base.Object(ctx, S3DriveOption).backend.(*S3Drive); !ok {
		t.Fatalf("Object(s3) backend = %T, want *S3Drive", base.Object(ctx, S3DriveOption).backend)
	}
	if got := base.Object(ctx, R2DriveOption); got.Option() != R2DriveOption {
		t.Fatalf("Object(r2).Option() = %q, want %q", got.Option(), R2DriveOption)
	}
	if _, ok := base.R2(ctx).backend.(*S3Drive); !ok {
		t.Fatalf("R2() backend = %T, want *S3Drive", base.R2(ctx).backend)
	}
}

func TestBaseDriveCloudUsesConfiguredProvider(t *testing.T) {
	ctx := context.Background()

	legacy := InitWithConfig(Config{S3: S3Config{Region: "eu-south-1"}})
	if got := legacy.Cloud(ctx); got.Option() != S3DriveOption {
		t.Fatalf("legacy Cloud().Option() = %q, want %q", got.Option(), S3DriveOption)
	}

	r2 := InitWithConfig(Config{S3: S3Config{
		Endpoint: "https://account-id.r2.cloudflarestorage.com",
	}})
	if got := r2.Cloud(ctx); got.Option() != R2DriveOption {
		t.Fatalf("r2 Cloud().Option() = %q, want %q", got.Option(), R2DriveOption)
	}

	explicitR2 := InitWithConfig(Config{S3: S3Config{Provider: "r2"}})
	if got := explicitR2.Cloud(ctx); got.Option() != R2DriveOption {
		t.Fatalf("explicit r2 Cloud().Option() = %q, want %q", got.Option(), R2DriveOption)
	}
}
