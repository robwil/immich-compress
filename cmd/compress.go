package cmd

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"immich-compress/compress"
	"immich-compress/immich"

	"github.com/spf13/cobra"
)

// formatSlice converts ImageFormat slice to string slice
func formatSlice[T ~string](formats []T) []string {
	result := make([]string, len(formats))
	for i, format := range formats {
		result[i] = string(format)
	}
	return result
}

var flagsCompress struct {
	flagParallel       int
	flagLimit          int
	flagDiff           int
	flagServer         string
	flagAPIKey         string
	flagAssetType      string
	flagAssetUUIDs     []string
	flagImageQuality   int
	flagImageEffort    int
	flagImageFormat    string
	flagVideoParallel  int
	flagVideoQuality   int
	flagVideoFormat    string
	flagVideoContainer string
	flagUserKeysFile   string
	flagDryRun         bool
	flagForce          bool
	flagMatchExtension string
	flagAfter          time.Time
	flagCreatedAfter   string
	flagCreatedBefore  string
	flagTakenAfter     string
	flagTakenBefore    string
}

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Compress photos and videos in an Immich library",
	Long: `Download assets from an Immich server, compress them using libvips (images)
or FFmpeg (videos), and re-upload the compressed versions. The original
assets are moved to trash after successful replacement.

Supports per-user API keys (--user-keys) for libraries with multiple users,
since Immich restricts asset operations to the asset owner.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var userKeys *immich.UserKeys
		if flagsCompress.flagUserKeysFile != "" {
			var err error
			userKeys, err = immich.LoadUserKeys(flagsCompress.flagUserKeysFile)
			if err != nil {
				return fmt.Errorf("failed to load user keys: %w", err)
			}
		}

		parseOptionalTime := func(s string) (*time.Time, error) {
			if s == "" {
				return nil, nil
			}
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return nil, fmt.Errorf("invalid date %q (expected YYYY-MM-DD): %w", s, err)
			}
			return &t, nil
		}

		createdAfter, err := parseOptionalTime(flagsCompress.flagCreatedAfter)
		if err != nil {
			return err
		}
		createdBefore, err := parseOptionalTime(flagsCompress.flagCreatedBefore)
		if err != nil {
			return err
		}
		takenAfter, err := parseOptionalTime(flagsCompress.flagTakenAfter)
		if err != nil {
			return err
		}
		takenBefore, err := parseOptionalTime(flagsCompress.flagTakenBefore)
		if err != nil {
			return err
		}

		var after *time.Time
		if cmd.Flags().Changed("after") {
			after = &flagsCompress.flagAfter
		}

		config := compress.Config{
			Parallel:       flagsCompress.flagParallel,
			Limit:          flagsCompress.flagLimit,
			AssetType:      flagsCompress.flagAssetType,
			AssetUUIDs:     flagsCompress.flagAssetUUIDs,
			Server:         flagsCompress.flagServer,
			APIKey:         flagsCompress.flagAPIKey,
			UserKeys:       userKeys,
			After:          after,
			CreatedAfter:   createdAfter,
			CreatedBefore:  createdBefore,
			TakenAfter:     takenAfter,
			TakenBefore:    takenBefore,
			DryRun:         flagsCompress.flagDryRun,
			Force:          flagsCompress.flagForce,
			MatchExtension: strings.ToLower(strings.TrimSpace(flagsCompress.flagMatchExtension)),
			DiffPercent:    flagsCompress.flagDiff,
			ImageQuality:   flagsCompress.flagImageQuality,
			ImageEffort:    flagsCompress.flagImageEffort,
			ImageFormat:    (compress.ImageFormat)(strings.ToLower(strings.TrimSpace(flagsCompress.flagImageFormat))),
			VideoContainer: (compress.VideoContainer)(strings.ToLower(strings.TrimSpace(flagsCompress.flagVideoContainer))),
			VideoFormat:    (compress.VideoFormat)(strings.ToLower(strings.TrimSpace(flagsCompress.flagVideoFormat))),
			VideoQuality:   flagsCompress.flagVideoQuality,
			VideoParallel:  flagsCompress.flagVideoParallel,
		}
		return compress.Compressing(cmd.Context(), config)
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	compressCmd.PersistentFlags().IntVarP(&flagsCompress.flagParallel, "parallel", "p", runtime.NumCPU(), "Number of parallel workers")
	compressCmd.PersistentFlags().IntVarP(&flagsCompress.flagLimit, "limit", "l", 0, "Maximum number of assets to compress (0 = no limit)")
	compressCmd.PersistentFlags().StringVarP(&flagsCompress.flagServer, "server", "s", "", "The immich server address")
	if err := compressCmd.MarkPersistentFlagRequired("server"); err != nil {
		panic(err)
	}
	compressCmd.PersistentFlags().StringVarP(&flagsCompress.flagAPIKey, "api-key", "a", "", "The immich server API key")
	if err := compressCmd.MarkPersistentFlagRequired("api-key"); err != nil {
		panic(err)
	}
	compressCmd.PersistentFlags().StringVarP(&flagsCompress.flagAssetType, "type", "i", "ALL", "Asset type to compress (IMAGE, VIDEO, ALL)")
	compressCmd.PersistentFlags().StringArrayVarP(&flagsCompress.flagAssetUUIDs, "uuid", "u", []string{}, "Asset UUID")
	compressCmd.PersistentFlags().IntVarP(&flagsCompress.flagImageQuality, "image-quality", "q", 80, "Image quality for compression (1-100)")
	compressCmd.PersistentFlags().IntVarP(&flagsCompress.flagImageEffort, "image-effort", "e", 7, "Image compression effort (1-9, higher = slower but smaller)")
	compressCmd.PersistentFlags().StringVarP(&flagsCompress.flagImageFormat, "image-format", "f", string(compress.JXL), fmt.Sprintf("Image format for compression (%v)", strings.Join(formatSlice(compress.ImageFormatsAvailable), ", ")))
	compressCmd.PersistentFlags().IntVar(&flagsCompress.flagVideoParallel, "video-parallel", 1, "Max concurrent video encodes (default 1, since ffmpeg is already multi-threaded)")
	compressCmd.PersistentFlags().IntVarP(&flagsCompress.flagVideoQuality, "video-quality", "Q", 25, "Video quality for compression (1-100). Lower is higher quality")
	compressCmd.PersistentFlags().StringVarP(&flagsCompress.flagVideoFormat, "video-format", "F", string(compress.AV1), fmt.Sprintf("Video format for compression (%v)", strings.Join(formatSlice(compress.VideoFormatsAvailable), ", ")))
	compressCmd.PersistentFlags().StringVarP(&flagsCompress.flagVideoContainer, "video-container", "C", string(compress.MKV), fmt.Sprintf("Video container format (%v)", strings.Join(formatSlice(compress.VideoContainersAvailable), ", ")))
	compressCmd.PersistentFlags().IntVarP(&flagsCompress.flagDiff, "diff-percents", "D", 8, "If size diff is lower than this percent files will not be replaced with new.")
	compressCmd.PersistentFlags().StringVar(&flagsCompress.flagUserKeysFile, "user-keys", "", "Path to JSON file mapping user IDs to API keys for cross-user asset deletion")
	compressCmd.PersistentFlags().BoolVarP(&flagsCompress.flagDryRun, "dry-run", "n", false, "Show matching assets without compressing")
	compressCmd.PersistentFlags().BoolVar(&flagsCompress.flagForce, "force", false, "Reprocess all matching assets, ignoring the compressed skip list")
	compressCmd.PersistentFlags().StringVar(&flagsCompress.flagMatchExtension, "match-ext", "", "Only process assets with this file extension (e.g. mkv, jpg)")
	compressCmd.PersistentFlags().TimeVarP(&flagsCompress.flagAfter, "after", "t", time.Now(), []string{"2006-01-02 15:04:05"}, "Skip assets already compressed after this time (recompression threshold)")
	compressCmd.PersistentFlags().StringVar(&flagsCompress.flagCreatedAfter, "created-after", "", "Only process assets uploaded after this date (YYYY-MM-DD)")
	compressCmd.PersistentFlags().StringVar(&flagsCompress.flagCreatedBefore, "created-before", "", "Only process assets uploaded before this date (YYYY-MM-DD)")
	compressCmd.PersistentFlags().StringVar(&flagsCompress.flagTakenAfter, "taken-after", "", "Only process assets captured after this date (YYYY-MM-DD)")
	compressCmd.PersistentFlags().StringVar(&flagsCompress.flagTakenBefore, "taken-before", "", "Only process assets captured before this date (YYYY-MM-DD)")
}
