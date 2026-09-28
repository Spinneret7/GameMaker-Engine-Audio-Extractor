# GameMaker-Engine-Audio-Extractor
Open Source Extractor for unpacking Audiofiles of GameMaker Engine like Audiogroup1.dat

### Compile:
```go build -o gme.exe main.go```

You must need go(Mandatory) and ffmpeg installed in your system for using this. However ffmpeg is optional and only needed if you would want to convert the pcm wave files to libvorbis!

### Usage
```
gme.exe [flags] <audiogroup.dat> [output folder]
  -raw          extract original bytes only (no WAV→Vorbis encode)
  -keep-raw     when encoding, also keep the original .wav/.ogg
  -q <n>        libvorbis quality for ffmpeg (default 4)
  -n <count>    only process first N tracks
Example: 1. gme.exe -raw "Audiogroup1.dat" "Extracted_Output"
2. gme.exe -q7 "Audiogroup1.dat" "Extracted_Output"
```
### Logs in Action:
```
umber of tracks: 28
File 001 at 0x00000084 (Size 31046444, WAV/PCM)... saved extract001.wav
File 002 at 0x01d9bbb4 (Size 35456444, WAV/PCM)... saved extract002.wav
File 003 at 0x03f6c174 (Size 32857836, WAV/PCM)... saved extract003.wav
File 004 at 0x05ec2064 (Size 22755644, WAV/PCM)... saved extract004.wav
File 005 at 0x074759a4 (Size 31367176, WAV/PCM)... saved extract005.wav
File 006 at 0x0925f9b0 (Size 36239504, WAV/PCM)... saved extract006.wav
File 007 at 0x0b4ef244 (Size 33774748, WAV/PCM)... saved extract007.wav
File 008 at 0x0d524ee4 (Size 40572044, WAV/PCM)... saved extract008.wav
File 009 at 0x0fbd6374 (Size 33163244, WAV/PCM)... saved extract009.wav
File 010 at 0x11b76b64 (Size 33868844, WAV/PCM)... saved extract010.wav
File 011 at 0x13bc3794 (Size 33806196, WAV/PCM)... saved extract011.wav
File 012 at 0x15c00f0c (Size 39949356, WAV/PCM)... saved extract012.wav
File 013 at 0x1821a33c (Size 32545844, WAV/PCM)... saved extract013.wav
File 014 at 0x1a123f74 (Size 32457644, WAV/PCM)... saved extract014.wav
File 015 at 0x1c018324 (Size 10584044, WAV/PCM)... saved extract015.wav
File 016 at 0x1ca30314 (Size 33979628, WAV/PCM)... saved extract016.wav
File 017 at 0x1ea98004 (Size 30937892, WAV/PCM)... saved extract017.wav
File 018 at 0x2081932c (Size 37168136, WAV/PCM)... saved extract018.wav
File 019 at 0x22b8b738 (Size 34574444, WAV/PCM)... saved extract019.wav
File 020 at 0x24c847a8 (Size 33078248, WAV/PCM)... saved extract020.wav
File 021 at 0x26c10394 (Size 34926404, WAV/PCM)... saved extract021.wav
File 022 at 0x28d5f2dc (Size 45985764, WAV/PCM)... saved extract022.wav
File 023 at 0x2b93a2c4 (Size 45985768, WAV/PCM)... saved extract023.wav
File 024 at 0x2e5152b0 (Size 20462444, WAV/PCM)... saved extract024.wav
File 025 at 0x2f898e20 (Size 27783044, WAV/PCM)... saved extract025.wav
File 026 at 0x31317da8 (Size 34221644, WAV/PCM)... saved extract026.wav
File 027 at 0x333babf8 (Size 32014060, WAV/PCM)... saved extract027.wav
File 028 at 0x35242ae8 (Size 32531796, WAV/PCM)... saved extract028.wav
Done: 28 tracks, 28 raw saved, 0 encoded to Vorbis.
Processed in 1.5064549s
```

This tool is dead simple to use, I forked it from [GMSAE by Jonathan Hecl](https://github.com/jonathanhecl/GMS-audiogroup-extractor) to export actual 16 bit pcm wave header instead of filename being fake ogg.
