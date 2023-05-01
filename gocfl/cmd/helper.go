package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/google/tink/go/core/registry"
	"github.com/je4/filesystem/v2/pkg/osfsrw"
	"github.com/je4/filesystem/v2/pkg/s3fsrw"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/filesystem/v2/pkg/zipfs"
	"github.com/je4/filesystem/v2/pkg/zipfsrw"
	defaultextensions_object "github.com/je4/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "github.com/je4/gocfl/v2/data/defaultextensions/storageroot"
	"github.com/je4/gocfl/v2/pkg/extension"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/pkg/subsystem/migration"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

func startTimer() *timer {
	t := &timer{}
	t.Start()
	return t
}

type timer struct {
	start time.Time
}

func (t *timer) Start() {
	t.start = time.Now()
}

func (t *timer) String() string {
	delta := time.Now().Sub(t.start)
	return delta.String()
}

func initExtensionFactory(extensionParams map[string]string, indexerAddr string, indexerActions *ironmaiden.ActionDispatcher, migration *migration.Migration, sourceFS fs.FS, logger *logging.Logger) (*ocfl.ExtensionFactory, error) {
	logger.Debugf("initializing ExtensionFactory")
	extensionFactory, err := ocfl.NewExtensionFactory(extensionParams, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension factory")
	}

	logger.Debugf("adding creator for extension %s", extension.DigestAlgorithmsName)
	extensionFactory.AddCreator(extension.DigestAlgorithmsName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDigestAlgorithmsFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutFlatDirectName)
	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutHashAndIdNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutHashedNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.FlatOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.FlatOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFlatOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.NTupleOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.NTupleOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewNTupleOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.DirectCleanName)
	extensionFactory.AddCreator(extension.DirectCleanName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.PathDirectName)
	extensionFactory.AddCreator(extension.PathDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutPairTreeName)
	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", ocfl.ExtensionManagerName)
	extensionFactory.AddCreator(ocfl.ExtensionManagerName, func(fsys fs.FS) (ocfl.Extension, error) {
		return ocfl.NewInitialDummyFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.ContentSubPathName)
	extensionFactory.AddCreator(extension.ContentSubPathName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewContentSubPathFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.MetaFileName)
	extensionFactory.AddCreator(extension.MetaFileName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMetaFileFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.IndexerName)
	extensionFactory.AddCreator(extension.IndexerName, func(fsys fs.FS) (ocfl.Extension, error) {
		ext, err := extension.NewIndexerFS(fsys, indexerAddr, indexerActions)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create new indexer from filesystem")
		}
		return ext, nil
	})

	logger.Debugf("adding creator for extension %s", extension.MigrationName)
	extensionFactory.AddCreator(extension.MigrationName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMigrationFS(fsys, migration)
	})

	logger.Debugf("adding creator for extension %s", extension.FilesystemName)
	extensionFactory.AddCreator(extension.FilesystemName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFilesystemFS(fsys)
	})

	return extensionFactory, nil
}

func GetExtensionParams() []*ocfl.ExtensionExternalParam {
	var result = []*ocfl.ExtensionExternalParam{}

	result = append(result, extension.GetIndexerParams()...)
	result = append(result, extension.GetMetaFileParams()...)
	result = append(result, extension.GetContentSubPathParams()...)

	return result
}

func GetExtensionParamValues(cmd *cobra.Command) map[string]string {
	var result = map[string]string{}
	extParams := GetExtensionParams()
	for _, param := range extParams {
		name, value := param.GetParam(cmd)
		if name != "" {
			result[name] = value
		}
	}
	return result
}

func initDefaultExtensions(extensionFactory *ocfl.ExtensionFactory, storageRootExtensionsFolder, objectExtensionsFolder string) (storageRootExtensions, objectExtensions []ocfl.Extension, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS, err = osfsrw.NewFS(storageRootExtensionsFolder)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for storage root extensions folder %v", storageRootExtensionsFolder)
		}
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS, err = osfsrw.NewFS(objectExtensionsFolder)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for object extensions folder %v", objectExtensionsFolder)
		}
	}
	storageRootExtensions, err = extensionFactory.LoadExtensions(dStorageRootExtDirFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dStorageRootExtDirFS)
		return
	}
	objectExtensions, err = extensionFactory.LoadExtensions(dObjectExtDirFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dObjectExtDirFS)
		return
	}
	return
}

func initializeFSFactory(prefix string, cmd *cobra.Command, zipDigests []checksum.DigestAlgorithm, readOnly bool, logger *logging.Logger) (*writefs.Factory, error) {
	if zipDigests == nil {
		zipDigests = []checksum.DigestAlgorithm{checksum.DigestSHA512}
	}
	prefix = strings.TrimRight(prefix, ".") + "."

	fsFactory, err := writefs.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create filesystem factory")
	}

	if readOnly {
		if err := fsFactory.Register(zipfs.NewCreateFSFunc(), "\\.zip$", writefs.HighFS); err != nil {
			return nil, errors.Wrap(err, "cannot register zipfs")
		}
	} else {

		flagNoCompression := viper.GetBool(prefix + "NoCompression")

		flagAES := viper.GetBool(prefix + "AES")

		keePassFile := viper.GetString(prefix + "KeePassFile")
		keePassEntry := viper.GetString(prefix + "KeePassEntry")
		keePassKey := viper.GetString(prefix + "KeePassKey")
		// todo: allow different KMS clients
		if flagAES {
			db, err := keepass2kms.LoadKeePassDBFromFile(keePassFile, keePassKey)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot load keepass file '%s'", keePassFile)
			}
			client, err := keepass2kms.NewClient(db, filepath.Base(keePassFile))
			if err != nil {
				return nil, errors.Wrap(err, "cannot create keepass2kms client")
			}
			registry.RegisterKMSClient(client)

			if err := fsFactory.Register(zipfsrw.NewCreateFSEncryptedChecksumFunc(flagNoCompression, zipDigests, keePassEntry), "\\.zip$", writefs.HighFS); err != nil {
				return nil, errors.Wrap(err, "cannot register FSEncryptedChecksum")
			}
		} else {
			if err := fsFactory.Register(zipfsrw.NewCreateFSChecksumFunc(flagNoCompression, zipDigests), "\\.zip$", writefs.HighFS); err != nil {
				return nil, errors.Wrap(err, "cannot register FSChecksum")
			}
		}
	}
	if err := fsFactory.Register(osfsrw.NewCreateFSFunc(), "", writefs.LowFS); err != nil {
		return nil, errors.Wrap(err, "cannot register osfs")
	}
	s3Endpoint := viper.GetString("S3Endpoint")
	s3AccessKeyID := viper.GetString("S3AccessKeyID")
	s3SecretAccessKey := viper.GetString("S3SecretAccessKey")
	if err := fsFactory.Register(
		s3fsrw.NewCreateFSFunc(
			map[string]*s3fsrw.S3Access{
				"switch": {
					s3AccessKeyID,
					s3SecretAccessKey,
					s3Endpoint,
					true,
				},
			},
			s3fsrw.ARNRegexStr,
			logger,
		),
		s3fsrw.ARNRegexStr,
		writefs.MediumFS,
	); err != nil {
		return nil, errors.Wrap(err, "cannot register s3fs")
	}

	return fsFactory, nil
}

func showStatus(ctx context.Context) error {
	status, err := ocfl.GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get status of validation")
	}
	status.Compact()
	contextString := ""
	errs := 0
	for _, err := range status.Errors {
		if err.Code[0] == 'E' {
			errs++
		}
		if err.Context != contextString {
			fmt.Printf("\n[%s]\n", err.Context)
			contextString = err.Context
		}
		fmt.Printf("   #%s - %s [%s]\n", err.Code, err.Description, err.Description2)
		//logger.Infof("ERROR: %v", err)
	}
	if errs > 0 {
		fmt.Printf("\n%d errors found\n", errs)
	} else {
		fmt.Printf("\nno errors found\n")
	}
	return nil
}

func addObjectByPath(
	storageRoot ocfl.StorageRoot,
	fixity []checksum.DigestAlgorithm,
	defaultExtensions []ocfl.Extension,
	checkDuplicates bool,
	id, userName, userAddress, message string,
	sourceFS fs.FS, area string,
	areaPaths map[string]fs.FS,
	echo bool) (bool, error) {
	var o ocfl.Object
	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		return false, errors.Wrapf(err, "cannot check for existence of %s", id)
	}
	if exists {
		o, err = storageRoot.LoadObjectByID(id)
		if err != nil {
			return false, errors.Wrapf(err, "cannot load object %s", id)
		}
	} else {
		o, err = storageRoot.CreateObject(id, storageRoot.GetVersion(), storageRoot.GetDigest(), fixity, defaultExtensions)
		if err != nil {
			return false, errors.Wrapf(err, "cannot create object %s", id)
		}
	}
	if err := o.StartUpdate(message, userName, userAddress, echo); err != nil {
		return false, errors.Wrapf(err, "cannot start update for object %s", id)
	}

	if err := o.AddFolder(sourceFS, checkDuplicates, area); err != nil {
		return false, errors.Wrapf(err, "cannot add folder '%s' to '%s'", sourceFS, id)
	}
	if areaPaths != nil {
		for a, aPath := range areaPaths {
			if err := o.AddFolder(aPath, checkDuplicates, a); err != nil {
				return false, errors.Wrapf(err, "cannot add area '%s' folder '%s' to '%s'", a, aPath, id)
			}
		}
	}
	if err := o.EndUpdate(); err != nil {
		return false, errors.Wrapf(err, "cannot end update for object '%s'", id)
	}

	if err := o.Close(); err != nil {
		return false, errors.Wrapf(err, "cannot close object '%s'", id)
	}

	return o.IsModified(), nil
}
