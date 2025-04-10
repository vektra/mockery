package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-errors/errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrNoNewVersion = errors.New("no new version specified")

var EXIT_CODE_NO_NEW_VERSION = 8

func NewTagCmd(v *viper.Viper) (*cobra.Command, error) {
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cmd := &cobra.Command{
		Use: "tag",
		Run: func(cmd *cobra.Command, args []string) {
			tagger, err := NewTagger(v)
			if err != nil {
				printStack(err)
				os.Exit(1)
			}
			requestedVersion, previousVersion, err := tagger.Tag()
			if requestedVersion != nil && previousVersion != nil {
				fmt.Fprintf(os.Stdout, "v%s,v%s", requestedVersion.String(), previousVersion.String())
			}
			if err != nil {
				if errors.Is(ErrNoNewVersion, err) {
					os.Exit(EXIT_CODE_NO_NEW_VERSION)
				}
				printStack(err)
				os.Exit(1)
			}
		},
	}
	flags := cmd.PersistentFlags()
	flags.Bool("dry-run", true, "print, but do not perform, any actions")

	viper.BindPFlag("dry-run", flags.Lookup("dry-run"))

	return cmd, nil
}

func (t *Tagger) createTag(repo *git.Repository, version string) error {
	hash, err := repo.Head()
	if err != nil {
		return errors.New(err)
	}

	if t.DryRun {
		logger.Info().Str("tag", version).Msg("would have created tag")
		return nil
	}
	majorVersion := strings.Split(version, ".")[0]
	for _, v := range []string{version, majorVersion} {
		if err := repo.DeleteTag(v); err != nil {
			logger.Warn().Err(err).Str("tag", v).Msg("failed to delete tag, might be okay.")
		}
		_, err = repo.CreateTag(v, hash.Hash(), &git.CreateTagOptions{
			Tagger: &object.Signature{
				Name:  "Landon Clipp",
				Email: "11232769+LandonTClipp@users.noreply.github.com",
				When:  time.Now(),
			},
			Message: v,
		})
		if err != nil {
			return errors.New(err)
		}
	}

	logger.Info().Str("tag", version).Msg("tag successfully created")
	return nil
}

func (t *Tagger) largestTagSemver(repo *git.Repository, major uint64) (*semver.Version, error) {
	largestTag, err := semver.NewVersion("v0.0.0")
	if err != nil {
		return nil, errors.New(err)
	}

	iter, err := repo.Tags()
	if err != nil {
		return nil, errors.New(err)
	}
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		var versionString string
		tag, err := repo.TagObject(ref.Hash())
		switch err {
		case nil:
		case plumbing.ErrObjectNotFound:
			// Not a tag object
		default:
			// Some other error
			return errors.New(err)
		}
		if err != nil {
			if errors.Is(plumbing.ErrObjectNotFound, err) {
				// Tag is lightweight tag
				versionString = ref.Name().Short()
			} else {
				logger.Err(err).
					Str("hash", ref.Hash().String()).
					Str("name", ref.Name().String()).
					Msg("error when retrieving tag object")
				return errors.New(err)
			}
		} else {
			versionString = tag.Name
		}
		versionParts := strings.Split(versionString, ".")
		if len(versionParts) < 3 {
			// This is not a full version tag, so ignore it
			return nil
		}

		version, err := semver.NewVersion(versionString)
		if err != nil {
			return errors.New(err)
		}
		if version.GreaterThan(largestTag) && version.Major() == major {
			largestTag = version
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return largestTag, nil
}

func NewTagger(v *viper.Viper) (*Tagger, error) {
	t := &Tagger{}
	if err := v.Unmarshal(t); err != nil {
		return nil, errors.New(err)
	}
	logger.Info().Msgf("Using config: %s", v.ConfigFileUsed())
	if err := validator.New(
		validator.WithRequiredStructEnabled(),
	).Struct(t); err != nil {
		return nil, errors.New(err)
	}
	return t, nil
}

type Tagger struct {
	DryRun  bool   `mapstructure:"dry-run"`
	Version string `mapstructure:"version" validate:"required"`
}

func (t *Tagger) Tag() (requestedVersion *semver.Version, previousVersion *semver.Version, err error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, nil, errors.New(err)
	}

	requestedVersion, err = semver.NewVersion(t.Version)
	if err != nil {
		logger.Err(err).Str("requested-version", string(t.Version)).Msg("error when constructing semver from version config")
		return requestedVersion, nil, errors.New(err)
	}

	previousVersion, err = t.largestTagSemver(repo, requestedVersion.Major())
	if err != nil {
		return requestedVersion, previousVersion, err
	}
	logger := logger.With().
		Stringer("previous-version", previousVersion).Logger()

	logger.Info().Msg("found largest semver tag")

	logger = logger.With().
		Stringer("requested-version", requestedVersion).
		Logger()
	if !requestedVersion.GreaterThan(previousVersion) {
		logger.Info().
			Msg("VERSION is not greater than latest git tag, nothing to do.")
		return requestedVersion, previousVersion, ErrNoNewVersion
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return requestedVersion, previousVersion, errors.New(err)
	}

	status, err := worktree.Status()
	if err != nil {
		return requestedVersion, previousVersion, errors.New(err)
	}
	if !status.IsClean() {
		logger.Error().Msg("git is in a dirty state, can't tag.")
		fmt.Println(status.String())
		return requestedVersion, previousVersion, errors.New("dirty git state")
	}

	if err := t.createTag(repo, fmt.Sprintf("v%s", requestedVersion.String())); err != nil {
		return requestedVersion, previousVersion, err
	}
	logger.Info().Msg("created new tag. Push to origin still required.")

	return requestedVersion, previousVersion, nil
}
