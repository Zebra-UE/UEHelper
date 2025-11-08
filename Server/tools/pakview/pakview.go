package pakview

import (
	"UEHelper/tools/misc"
	"errors"
	"fmt"
	"os"
	"strings"
)

type FMacro struct {
	ENABLE_PAKFILE_RUNTIME_PRUNING int32
}

func InitMacro() FMacro {
	var result FMacro
	result.ENABLE_PAKFILE_RUNTIME_PRUNING = 1

	return result
}

var Macro FMacro

const (
	PakFile_Version_Initial = 1

	PakFile_Version_CompressionEncryption       = 3
	PakFile_Version_IndexEncryption             = 4
	PakFile_Version_EncryptionKeyGuid           = 7
	PakFile_Version_FNameBasedCompressionMethod = 8
	PakFile_Version_FrozenIndex                 = 9
	PakFile_Version_PathHashIndex               = 10
	PakFile_Version_Last                        = 12
	PakFile_Version_Latest                      = PakFile_Version_Last - 1
)

const (
	PakFile_Magic            = 0x5A6F12E1
	MaxNumCompressionMethods = 4
	CompressionMethodNameLen = 32
)

const (
	COMPRESS_None              = 0
	COMPRESS_ZLIB_DEPRECATED   = 0x01
	COMPRESS_GZIP_DEPRECATED   = 0x02
	COMPRESS_Custom_DEPRECATED = 0x04
)

func getSerializeSize(version int) int64 {
	var result int64 = 45

	if version >= PakFile_Version_EncryptionKeyGuid {
		result += 16
	}
	if version >= PakFile_Version_FNameBasedCompressionMethod {
		result += (32 * 5)
	}
	if version >= PakFile_Version_FrozenIndex && version <= PakFile_Version_PathHashIndex {
		result += 1
	}

	return result
}

type Numerical interface {
	int8 | uint8 | int32 | uint32 | int64 | uint64
}

func bytesToNumerical[T Numerical](chunk []byte) T {
	var result T = 0
	len := len(chunk)
	for i := len - 1; i >= 0; i-- {
		result = result<<8 + T(chunk[i])
	}
	return result
}

func bytesToString(chunk []byte) string {
	for idx, c := range chunk {
		if c == 0 {
			return string(chunk[:idx])
		}
	}
	return string(chunk)
}

func readBool(reader FArchive) bool {

	var chunk [4]byte
	reader.Read(chunk[:])
	return chunk[0] > 0

}

func readNumerical[T Numerical](reader FArchive, len int) T {
	var chunk = make([]byte, len)
	n, err := reader.Read(chunk[:])
	if err != nil || n != len {
		return 0
	}
	var result T = 0
	for i := len - 1; i >= 0; i-- {
		result = result<<8 + T(chunk[i])
	}
	return result
}

type FGuid struct {
	V [4]uint32
}

func readGUID(reader FArchive, result *FGuid) {

	for i := 0; i < 4; i++ {
		result.V[i] = readNumerical[uint32](reader, 4)
	}

}

func readString(reader FArchive) string {
	pos, _ := reader.Seek(0, 1)
	fmt.Print(pos)
	len := readNumerical[int32](reader, 4)

	chunk := make([]byte, len)

	reader.Read(chunk)

	return string(chunk[:len-1])

}

type FSHAHash struct {
	Hash [20]byte
}

func (me *FGuid) Invalidate() {
	for i := 0; i < 4; i++ {
		me.V[i] = 0
	}
}

func (me *FGuid) IsValid() bool {
	return (me.V[0] | me.V[1] | me.V[2] | me.V[3]) != 0
}

type FArchive interface {
	Read([]byte) (n int, err error)
	Seek(int64, int) (n int64, err error)
	TotalSize() int64
}

type FFileReader struct {
	innder *os.File
	size   int64
}

func createFileReader(file *os.File) *FFileReader {
	result := new(FFileReader)
	result.innder = file
	stat, err := file.Stat()
	if err != nil {
		return nil
	}
	result.size = stat.Size()
	return result
}

func (me *FFileReader) Read(b []byte) (int, error) {
	result, err := me.innder.Read(b)
	return result, err
}

func (me *FFileReader) Seek(offset int64, from int) (int64, error) {
	result, err := me.innder.Seek(offset, int(from))

	return result, err
}

func (me *FFileReader) TotalSize() int64 {
	return me.size
}

type FMemoryReader struct {
	data  []byte
	index int32
}

func (me *FMemoryReader) Read(chunk []byte) (int, error) {
	chunkLen := int32(len(chunk))
	copy(chunk, me.data[me.index:me.index+chunkLen])
	me.index += chunkLen
	return int(chunkLen), nil
}

func (me *FMemoryReader) Seek(offset int64, from int) (int64, error) {
	me.index += int32(offset)
	return int64(me.index), nil
}

func (me *FMemoryReader) TotalSize() int64 {
	return int64(len(me.data))
}

type FPakInfo struct {
	Version            int32
	IndexOffset        int64
	IndexSize          int64
	IndexHash          FSHAHash
	bEncryptedIndex    bool
	EncryptionKeyGuid  FGuid
	CompressionMethods [4]string
}

type FPakFile struct {
	NumEntries   int32
	PathHashSeed uint64

	directoryIndex           *FDirectoryIndex
	prunedDirectoryIndex     *FDirectoryIndex
	bWillPruneDirectoryIndex bool
	bNeedsLegacyPruning      bool
	bSomePakNeedsPruning     bool
}

type FPakCompressedBlock struct {
	compressedStart int64
	compressedEnd   int64
}

type FPakDirectoryRange struct {
	index int32
	count int32
}
type FPakEntryLocation struct {
	name  string
	index int32
}

type FDirectoryIndex struct {
	directory     map[string]FPakDirectoryRange
	entryLocation []FPakEntryLocation
}

func readDirectoryIndex(reader FArchive) *FDirectoryIndex {
	var result FDirectoryIndex
	directoryNum := readNumerical[int32](reader, 4)
	for i := int32(0); i < directoryNum; i++ {
		key := readString(reader)
		var directoryRanage FPakDirectoryRange
		directoryRanage.index = -1
		directoryRanage.count = 0
		fileNum := readNumerical[int32](reader, 4)
		if fileNum > 0 {
			directoryRanage.index = int32(len(result.entryLocation))
			for j := int32(0); j < fileNum; j++ {
				var entry FPakEntryLocation
				entry.name = readString(reader)
				entry.index = readNumerical[int32](reader, 4)
				result.entryLocation = append(result.entryLocation, entry)
			}
			directoryRanage.count = fileNum
		}
		result.directory[key] = directoryRanage

	}
	return &result
}

type FPakEntry struct {
	offset                 int64
	size                   int64
	uncompressedSize       int64
	flag                   uint8
	hash                   [20]uint8
	compressionMethodIndex uint32
	compressionBlockSize   uint32
	compressionBlocks      []FPakCompressedBlock
}

func (me *FPakEntry) Serialize(reader FArchive, version int32) {
	me.offset = readNumerical[int64](reader, 8)
	me.size = readNumerical[int64](reader, 8)
	me.uncompressedSize = readNumerical[int64](reader, 8)
	if version < PakFile_Version_FNameBasedCompressionMethod {
		var LegacyCompressionMethod int32
		LegacyCompressionMethod = readNumerical[int32](reader, 4)

		if LegacyCompressionMethod == COMPRESS_None {
			me.compressionMethodIndex = 0
		} else if LegacyCompressionMethod&COMPRESS_ZLIB_DEPRECATED > 0 {
			me.compressionMethodIndex = 1
		} else if LegacyCompressionMethod&COMPRESS_GZIP_DEPRECATED > 0 {
			me.compressionMethodIndex = 2
		} else if LegacyCompressionMethod&COMPRESS_Custom_DEPRECATED > 0 {
			me.compressionMethodIndex = 3
		}
	} else {
		me.compressionMethodIndex = readNumerical[uint32](reader, 4)
	}
	if version <= PakFile_Version_Initial {
		ticks := readNumerical[int64](reader, 8)
		fmt.Print(ticks)
	}
	reader.Read(me.hash[:])
	if version >= PakFile_Version_CompressionEncryption {
		if me.compressionMethodIndex != 0 {
			compressionBlockNum := int(readNumerical[int32](reader, 4))
			me.compressionBlocks = make([]FPakCompressedBlock, compressionBlockNum)
			for i := 0; i < compressionBlockNum; i++ {
				var block FPakCompressedBlock
				block.compressedStart = readNumerical[int64](reader, 8)
				block.compressedEnd = readNumerical[int64](reader, 8)
				me.compressionBlocks = append(me.compressionBlocks, block)
			}
		}
		me.flag = readNumerical[uint8](reader, 1)
		me.compressionBlockSize = readNumerical[uint32](reader, 4)
	}
}

func LoadIndex(reader FArchive, pakInfo FPakInfo) (*FPakFile, error) {
	//currentPos, _ := reader.Seek(0, 1)
	var pakFile FPakFile

	if pakInfo.Version >= PakFile_Version_PathHashIndex {
		pos, err := reader.Seek(pakInfo.IndexOffset, 0)
		if err != nil {
			return nil, errors.New("")
		}
		if pos != pakInfo.IndexOffset {
			return
		}
		PrimaryIndexData := make([]byte, pakInfo.IndexSize)
		aesKey := []byte{0xca, 0x9f, 0x79, 0x24, 0x9b, 0x5c, 0xf7, 0x79, 0x6, 0x96, 0x9e, 0x43, 0x99, 0x86, 0x11, 0xfd, 0x52, 0xd1, 0x9d, 0xce, 0x55, 0x4f, 0x50, 0x4b, 0x55, 0x73, 0xbf, 0x46, 0x5, 0x4f, 0x39, 0x9b}
		reader.Read(PrimaryIndexData)
		// Decrypt
		{
			if pakInfo.bEncryptedIndex {

				//aesKey := []byte("yp95JJtc93kGlp5DmYYR/VLRnc5VT1BLVXO/RgVPOZs=")
				PrimaryIndexData, err = misc.Decrypt(PrimaryIndexData, aesKey)
				if err != nil {

				}
			}

			//reader.Seek(pakInfo.IndexOffset, 0)
		}
		var memoryReader FMemoryReader
		memoryReader.data = PrimaryIndexData
		//memoryReader = &FMemoryReader({data:Pri})
		MountPoint := readString(&memoryReader)
		if len(MountPoint) > 0 && !strings.HasSuffix(MountPoint, "/") {
			MountPoint = MountPoint + "/"
		}

		pakFile.NumEntries = readNumerical[int32](&memoryReader, 4)

		pakFile.PathHashSeed = readNumerical[uint64](&memoryReader, 8)

		var bReaderHasPathHashIndex bool
		var PathHashIndexOffset int64
		var PathHashIndexSize int64
		var PathHashIndexHash FSHAHash

		bReaderHasPathHashIndex = readBool(&memoryReader)
		if bReaderHasPathHashIndex {
			PathHashIndexOffset = readNumerical[int64](&memoryReader, 8)
			PathHashIndexSize = readNumerical[int64](&memoryReader, 8)
			memoryReader.Read(PathHashIndexHash.Hash[:])
			bReaderHasPathHashIndex = PathHashIndexOffset >= 0
		}

		var bReaderHasFullDirectoryIndex bool
		var FullDirectoryIndexOffset int64
		var FullDirectoryIndexSize int64
		var FullDirectoryIndexHash FSHAHash

		bReaderHasFullDirectoryIndex = readBool(&memoryReader)
		if bReaderHasFullDirectoryIndex {
			FullDirectoryIndexOffset = readNumerical[int64](&memoryReader, 8)
			FullDirectoryIndexSize = readNumerical[int64](&memoryReader, 8)
			memoryReader.Read(FullDirectoryIndexHash.Hash[:])
			bReaderHasFullDirectoryIndex = FullDirectoryIndexOffset >= 0
		}
		fmt.Print(FullDirectoryIndexOffset)
		fmt.Print(FullDirectoryIndexSize)
		//EncodedPakEntries
		var EncodedPakEntriesNum int32
		EncodedPakEntriesNum = readNumerical[int32](&memoryReader, 4)
		EncodedPakEntries := make([]byte, EncodedPakEntriesNum)
		memoryReader.Read(EncodedPakEntries)

		var FilesNum int32
		FilesNum = readNumerical[int32](&memoryReader, 4)
		if FilesNum < 0 {
			return
		}

		Files := make([]FPakEntry, FilesNum)

		for i := 0; i < int(FilesNum); i++ {
			Files[i].Serialize(&memoryReader, pakInfo.Version)
		}

		var bWillUseFullDirectoryIndex bool
		var bWillUsePathHashIndex bool
		var bReadFullDirectoryIndex bool

		if bReaderHasPathHashIndex && bReaderHasFullDirectoryIndex {
			bWillUseFullDirectoryIndex = true
			bWillUsePathHashIndex = !bWillUseFullDirectoryIndex
			bReadFullDirectoryIndex = true
		} else if bReaderHasPathHashIndex {
			bWillUsePathHashIndex = true
			bWillUseFullDirectoryIndex = false
			bReadFullDirectoryIndex = false
		} else if bReaderHasFullDirectoryIndex {
			bWillUsePathHashIndex = false
			bWillUseFullDirectoryIndex = true
			bReadFullDirectoryIndex = true
		} else {
			return
		}

		if bWillUsePathHashIndex {
			PathHashIndexEndPosition := PathHashIndexOffset + PathHashIndexSize

			if PathHashIndexEndPosition < 100 {

			}
			reader.Seek(PathHashIndexOffset, 0)
			PathHashIndexData := make([]byte, PathHashIndexSize)
			reader.Read(PathHashIndexData)

			PathHashIndexData, _ = misc.Decrypt(PathHashIndexData, aesKey)
			var PathHashIndexReader FMemoryReader
			PathHashIndexReader.data = PathHashIndexData

		}

		if !bReadFullDirectoryIndex {

		} else {
			reader.Seek(FullDirectoryIndexOffset, 0)
			FullDirectoryIndexData := make([]byte, FullDirectoryIndexSize)
			reader.Read(FullDirectoryIndexData)

			FullDirectoryIndexData, err = misc.Decrypt(FullDirectoryIndexData, aesKey)
			if err != nil {

			}
			var SecondaryIndexReader FMemoryReader
			SecondaryIndexReader.data = FullDirectoryIndexData
			pakFile.directoryIndex = readDirectoryIndex(&SecondaryIndexReader)
		}

		if Macro.ENABLE_PAKFILE_RUNTIME_PRUNING == 1 {
			if bWillUseFullDirectoryIndex {
				pakFile.bWillPruneDirectoryIndex = false
			} else {
				pakFile.prunedDirectoryIndex = readDirectoryIndex(reader)
				pakFile.bWillPruneDirectoryIndex = true
				pakFile.bSomePakNeedsPruning = true
			}

		}

	} else {
		//LoadLegacyIndex
	}

	return &pakFile, nil
}

func loadPakInfo(reader FArchive) *FPakInfo {
	var result FPakInfo

	CompatibleVersion := PakFile_Version_Latest

	totalSize := reader.TotalSize()

	for {
		if CompatibleVersion < PakFile_Version_Initial {
			break
		}
		fileInfoPos := totalSize - getSerializeSize(CompatibleVersion)

		reader.Seek(fileInfoPos, 0)

		if CompatibleVersion >= PakFile_Version_EncryptionKeyGuid {
			readGUID(reader, &result.EncryptionKeyGuid)
		}
		var chunk [5]byte
		reader.Read(chunk[:])
		result.bEncryptedIndex = chunk[0] > 0

		magic := bytesToNumerical[int32](chunk[1:])
		if magic == PakFile_Magic {
			break
		}
		CompatibleVersion -= 1
	}

	result.Version = readNumerical[int32](reader, 4)
	result.IndexOffset = readNumerical[int64](reader, 8)
	result.IndexSize = readNumerical[int64](reader, 8)
	reader.Read(result.IndexHash.Hash[:])

	if result.Version < PakFile_Version_IndexEncryption {
		result.bEncryptedIndex = false
	}
	if result.Version < PakFile_Version_EncryptionKeyGuid {
		result.EncryptionKeyGuid.Invalidate()
	}
	if result.Version >= PakFile_Version_FrozenIndex && result.Version < PakFile_Version_PathHashIndex {
		var bIndexIsFrozen bool = false
		bIndexIsFrozen = readNumerical[uint8](reader, 1) > 0
		if bIndexIsFrozen {
			return nil
		}
	}

	if result.Version < PakFile_Version_FNameBasedCompressionMethod {
		result.CompressionMethods[0] = "Zlib"
		result.CompressionMethods[0] = "Gzip"
		result.CompressionMethods[0] = "Oodle"
	} else {
		bufferSize := CompressionMethodNameLen * MaxNumCompressionMethods
		methods := make([]byte, bufferSize)
		reader.Read(methods)
		for i := 0; i < MaxNumCompressionMethods; i++ {
			result.CompressionMethods[i] = bytesToString(methods[i*CompressionMethodNameLen : (i+1)*CompressionMethodNameLen])
		}
	}

	return &result
}

func Load(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	fileReader := createFileReader(file)
	pakInfo := loadPakInfo(fileReader)
	var pakFile *FPakFile
	if !pakInfo.EncryptionKeyGuid.IsValid() {
		pakFile = LoadIndex(fileReader, *pakInfo)
	}

}
