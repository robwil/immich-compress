// Package compress
package compress

import (
	"context"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	"immich-compress/immich"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Config holds configuration for compression
type Config struct {
	Parallel       int
	Limit          int
	AssetType      string
	AssetUUIDs     []string
	Server         string
	APIKey         string
	UserKeys       *immich.UserKeys
	DryRun         bool
	After          *time.Time
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	TakenAfter     *time.Time
	TakenBefore    *time.Time
	DiffPercent    int
	ImageFormat    ImageFormat
	ImageQuality   int
	ImageEffort    int
	VideoContainer VideoContainer
	VideoFormat    VideoFormat
	VideoQuality   int
	VideoParallel  int
}

func Compressing(ctx context.Context, config Config) error {
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(config.Parallel)
	client, err := immich.NewClientSimple(gCtx, config.Parallel, config.Server, config.APIKey, config.UserKeys)
	if err != nil {
		return err
	}

	videoSem := make(chan struct{}, config.VideoParallel)

	compressedIDs, err := client.CompressedAssetIDsAllUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch compressed assets: %w", err)
	}
	if len(compressedIDs) > 0 {
		fmt.Printf("Skipping %d already-compressed assets\n", len(compressedIDs))
	}

	var counter int32 = 0

	searchOption := immich.SearchAssetsJSONRequestBody{
		CreatedAfter:  config.CreatedAfter,
		CreatedBefore: config.CreatedBefore,
		TakenAfter:    config.TakenAfter,
		TakenBefore:   config.TakenBefore,
	}
	if config.AssetType != "ALL" {
		typeAsset := (immich.AssetTypeEnum)(config.AssetType)
		searchOption.Type = &typeAsset
	}
	if len(config.AssetUUIDs) == 1 {
		UUIDstring := config.AssetUUIDs[0]
		UUID, err := uuid.Parse(UUIDstring)
		if err != nil {
			return err
		}
		searchOption.Id = &UUID
	}
	ch := client.AssetSearch(config.Limit, searchOption)
	// Start the workers. Instead of 'for range parallel', we simply
	// read from the channel and run g.Go() for *each* element.
	// SetLimit(parallel) will take care of the limit.
	for asset := range ch {
		// Pass 'asset' to the closure to avoid race conditions
		asset := asset

		g.Go(func() error {
			// Check for cancellation. gCtx.Done() will trigger if:
			// 1. Parent 'ctx' is cancelled.
			// 2. Another goroutine in 'g' returned an error.
			select {
			case <-gCtx.Done():
				return gCtx.Err() // Return cancellation error
			default:
			}

			if asset.Err != nil {
				// Just return the error.
				// errgroup will automatically call cancel() for gCtx.
				return asset.Err
			}

			if len(config.AssetUUIDs) > 0 {
				if !slices.Contains(config.AssetUUIDs, asset.Asset.Id) {
					return nil
				}
			}

			if compressedIDs[asset.Asset.Id] {
				if config.After == nil || asset.Asset.FileModifiedAt.After(*config.After) {
					return nil
				}
			}

			atomic.AddInt32(&counter, 1)

			if config.DryRun {
				fmt.Printf("[dry-run] %s (%s, %s)\n", asset.Asset.OriginalFileName, asset.Asset.Id, asset.Asset.Type)
				return nil
			}

			// Process the asset here
			fmt.Printf("Processing file: %#v\n", asset.Asset.Id)
			err := compressFile(gCtx, client, asset.Asset, config.DiffPercent, videoSem, ImageConfig{
				Format:  config.ImageFormat,
				Quality: config.ImageQuality,
				Effort:  config.ImageEffort,
			}, VideoConfig{
				Container: config.VideoContainer,
				Format:    config.VideoFormat,
				Quality:   config.VideoQuality,
			})
			if err != nil {
				return err
			}

			return nil
		})
	}

	// g.Wait() waits for all goroutines to complete (like wg.Wait())
	// and returns the FIRST non-zero error returned by
	// any of the goroutines.
	if err := g.Wait(); err != nil {
		// If there was an error (including cancellation), we return it
		return err
	}

	if config.DryRun {
		fmt.Printf("Found %d assets to compress\n", counter)
	} else {
		fmt.Printf("Processed files: %d\n", counter)
	}

	return nil
}
