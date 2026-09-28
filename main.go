package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var version string = "1.2.0 ~ 09/28/2026"

func main() {
	fmt.Println("GameMaker Studio AudioGroup Extractor v" + version)
	fmt.Println("Author: Jonathan Hecl ~ https://www.jonathanhecl.com")
	fmt.Println("OGGRE-ready fork: auto-detects WAV/OGG and can encode Vorbis for OGGRE")
	fmt.Println()

	audiogroup := "audiogroup1.dat"
	outputFolder := "."
	convertToOgg := true // default: produce real Vorbis .ogg (needed by OGGRE)
	keepRaw := false
	vorbisQuality := "4" // ffmpeg libvorbis -q:a (approx -1..10)
	maxTracks := 0       // 0 = all

	args := os.Args[1:]
	positional := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			printUsage()
			return
		case a == "-raw":
			convertToOgg = false
		case a == "-keep-raw":
			keepRaw = true
		case strings.HasPrefix(a, "-q="):
			vorbisQuality = strings.TrimPrefix(a, "-q=")
		case a == "-q" && i+1 < len(args):
			i++
			vorbisQuality = args[i]
		case strings.HasPrefix(a, "-n="):
			maxTracks, _ = strconv.Atoi(strings.TrimPrefix(a, "-n="))
		case a == "-n" && i+1 < len(args):
			i++
			maxTracks, _ = strconv.Atoi(args[i])
		case strings.HasPrefix(a, "-"):
			fmt.Println("Unknown flag:", a)
			printUsage()
			return
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) >= 1 && positional[0] != "" {
		audiogroup = positional[0]
	}
	if len(positional) >= 2 && positional[1] != "" {
		outputFolder = positional[1]
	}

	fmt.Println("USAGE: " + filepath.Base(os.Args[0]) + " [flags] <audiogroup.dat> [output folder]")
	fmt.Println("  -raw          extract original bytes only (no WAV→Vorbis encode)")
	fmt.Println("  -keep-raw     when encoding, also keep the original .wav/.ogg")
	fmt.Println("  -q <n>        libvorbis quality for ffmpeg (default 4)")
	fmt.Println("  -n <count>    only process first N tracks")
	fmt.Println()

	started := time.Now()

	if _, err := os.Stat(audiogroup); err != nil {
		fmt.Println(audiogroup, "not found.")
		return
	}

	if outputFolder != "." {
		if err := os.MkdirAll(outputFolder, 0755); err != nil {
			fmt.Println("Could not create output folder:", err)
			return
		}
	}

	ffmpegPath := ""
	if convertToOgg {
		var err error
		ffmpegPath, err = exec.LookPath("ffmpeg")
		if err != nil {
			fmt.Println("WARNING: ffmpeg not found in PATH.")
			fmt.Println("         WAV tracks will be saved as .wav (OGGRE needs real Vorbis .ogg).")
			fmt.Println("         Install ffmpeg, or re-run after adding it to PATH.")
			convertToOgg = false
		}
	}

	fmt.Println("Processing", audiogroup, "...")
	fmt.Println("Output folder:", outputFolder)
	if convertToOgg {
		fmt.Println("Mode: encode WAV → Vorbis OGG (OGGRE-ready), quality", vorbisQuality)
	} else {
		fmt.Println("Mode: raw extract (correct extension by magic)")
	}

	f, err := os.Open(audiogroup)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bytesFile, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	fileSize := int64(len(bytesFile))
	fmt.Println("File size:", fileSize, "bytes")

	if fileSize < 0x18 || string(bytesFile[0:4]) != "FORM" {
		fmt.Println("Not a GameMaker FORM file.")
		return
	}

	tracks := int(binary.LittleEndian.Uint32(bytesFile[0x10:0x14]))
	if tracks <= 0 {
		fmt.Println("No tracks found.")
		return
	}
	if 0x14+tracks*4 > len(bytesFile) {
		fmt.Println("Track offset table extends past end of file.")
		return
	}

	fmt.Printf("Number of tracks: %d\n", tracks)
	if maxTracks > 0 && maxTracks < tracks {
		fmt.Printf("Limiting to first %d tracks (-n)\n", maxTracks)
		tracks = maxTracks
	}

	encoded := 0
	rawSaved := 0
	for n := 0; n < tracks; n++ {
		entryOff := int(binary.LittleEndian.Uint32(bytesFile[0x14+n*4 : 0x14+n*4+4]))
		if entryOff < 0 || entryOff+4 > len(bytesFile) {
			fmt.Printf("Track %03d: bad offset 0x%08x, skipping\n", n+1, entryOff)
			continue
		}
		trackSize := int(binary.LittleEndian.Uint32(bytesFile[entryOff : entryOff+4]))
		dataOff := entryOff + 4
		if trackSize < 0 || dataOff+trackSize > len(bytesFile) {
			fmt.Printf("Track %03d: bad size %d at 0x%08x, skipping\n", n+1, trackSize, entryOff)
			continue
		}
		data := bytesFile[dataOff : dataOff+trackSize]
		kind := detectAudio(data)
		ext := extensionFor(kind)

		fmt.Printf("File %03d at 0x%08x (Size %d, %s)... ", n+1, entryOff, trackSize, kind)

		rawName := fmt.Sprintf("extract%03d%s", n+1, ext)
		rawPath := filepath.Join(outputFolder, rawName)

		// Write embedded bytes when: raw mode, already OGG, or WAV that we must feed to ffmpeg.
		if err := os.WriteFile(rawPath, data, 0644); err != nil {
			panic(err)
		}
		rawSaved++

		if convertToOgg && kind == kindWAV {
			oggPath := filepath.Join(outputFolder, fmt.Sprintf("extract%03d.ogg", n+1))
			if err := encodeVorbis(ffmpegPath, rawPath, oggPath, vorbisQuality); err != nil {
				fmt.Printf("ffmpeg failed: %v\n", err)
				continue
			}
			encoded++
			if !keepRaw {
				_ = os.Remove(rawPath)
				rawSaved--
			}
			info, _ := os.Stat(oggPath)
			outSize := int64(0)
			if info != nil {
				outSize = info.Size()
			}
			fmt.Printf("-> %s (%d bytes, %.1f%% of wav)\n", filepath.Base(oggPath), outSize, 100*float64(outSize)/float64(trackSize))
		} else if kind == kindOGG {
			fmt.Printf("already Ogg, kept %s\n", rawName)
		} else {
			fmt.Printf("saved %s\n", rawName)
		}
	}

	fmt.Printf("Done: %d tracks, %d raw saved, %d encoded to Vorbis.\n", tracks, rawSaved, encoded)
	fmt.Println("Processed in", time.Since(started).String())
	if encoded > 0 {
		fmt.Println("Tip: run OGGRE on the .ogg files, e.g.")
		fmt.Println("  OGGRE_enc.exe Output\\extract001.ogg Output\\extract001.ogr")
	}
}

func printUsage() {
	fmt.Println("USAGE: " + filepath.Base(os.Args[0]) + " [flags] <audiogroup.dat> [output folder]")
	fmt.Println("Flags:")
	fmt.Println("  -raw          extract original embedded audio only (WAV stays WAV)")
	fmt.Println("  -keep-raw     keep original .wav alongside encoded .ogg")
	fmt.Println("  -q <n>        Vorbis quality for ffmpeg libvorbis (default 4)")
	fmt.Println("  -n <count>    only process first N tracks")
	fmt.Println()
	fmt.Println("By default, embedded WAV tracks are encoded to real Ogg Vorbis so OGGRE can")
	fmt.Println("detect codebooks and compress them. The stock extractor always named files")
	fmt.Println(".ogg even when the payload was uncompressed RIFF/WAV — OGGRE then reports")
	fmt.Println("~100% ratio and 0 codebooks.")
}

type audioKind int

const (
	kindUnknown audioKind = iota
	kindWAV
	kindOGG
)

func (k audioKind) String() string {
	switch k {
	case kindWAV:
		return "WAV/PCM"
	case kindOGG:
		return "Ogg"
	default:
		return "unknown"
	}
}

func detectAudio(data []byte) audioKind {
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte("OggS")) {
		return kindOGG
	}
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WAVE")) {
		return kindWAV
	}
	if len(data) >= 4 && bytes.Equal(data[0:4], []byte("RIFF")) {
		return kindWAV
	}
	return kindUnknown
}

func extensionFor(k audioKind) string {
	switch k {
	case kindWAV:
		return ".wav"
	case kindOGG:
		return ".ogg"
	default:
		return ".bin"
	}
}

func encodeVorbis(ffmpegPath, inPath, outPath, quality string) error {
	if _, err := strconv.ParseFloat(quality, 64); err != nil {
		return fmt.Errorf("invalid quality %q", quality)
	}
	cmd := exec.Command(ffmpegPath,
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", inPath,
		"-c:a", "libvorbis",
		"-q:a", quality,
		outPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
