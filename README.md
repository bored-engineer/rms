# rms - Go libraries for Azure Rights Management
`rms` is a MVP implementation of the necessary code to interact and decrypt (and maybe someday encrypt) Azure Rights Management protected content.
It is written in Go and has no dependencies on native libraries

## Should I use this in production?
Absolutely no. Under any circumstances. This library is an MVP state and was mainly written out of curiosity.

## Installation
```
go install github.com/bored-engineer/rms
```

## Usage
Generally this library is intended to be used via the API, however a `rms` CLI is bundled with the client for basic operations/debugging.

### Decrypt an rpmsg file
Decode the [rpmsg file|https://en.wikipedia.org/wiki/Rpmsg] into a [compound file|https://en.wikipedia.org/wiki/Compound_File_Binary_Format]:
```
$ rms rpmsg decode message.rpmsg
Decoded 52224 bytes to compound file: rpmsg.compound
```
Unpack the raw compound file:
```
$ rms compound unpack rpmsg.compound
Skipping empty entry: DataSpaces
Wrote 76 bytes from entry: DataSpaces/Version
Wrote 80 bytes from entry: DataSpaces/DataSpaceMap
Skipping empty entry: DataSpaces/DataSpaceInfo
Wrote 40 bytes from entry: DataSpaces/DataSpaceInfo/DRMDataSpace
Skipping empty entry: DataSpaces/TransformInfo
Skipping empty entry: DataSpaces/TransformInfo/DRMTransform
Wrote 43644 bytes from entry: DataSpaces/TransformInfo/DRMTransform/Primary
Wrote 4104 bytes from entry: DRMContent
Unpacked rpmsg.compound to ./unpacked/
```
Fetch a user license for the file:
```
$ rms license fetch "$access_token" unpacked/DataSpaces/TransformInfo/DRMTransform/Primary
{
	"Id": "00000000-0000-0000-0000-000000000000",
	...
	"AccessStatus": "AccessGranted",
	"Policy": {
		"AllowAuditedExtraction": true,
		"UserRights": [
			...
```
Decrypt the DRM content using the user license file:
```
$ rms license decrypt -o decrypted/ user.license unpacked/DRMContent
Decrypted 4096 bytes from unpacked/DRMContent
```
Unpack the decrypted DRM contents:
```
$ rms compound unpack decrypted.compound
Wrote 1554 bytes from entry: BodyPT-HTML
Wrote 16 bytes from entry: RpmsgStorageInfo
Wrote 6 bytes from entry: OutlookBodyStreamInfo
Unpacked decrypted.compound to ./decrypted/
```

## Prior Art
* https://www.usenix.org/system/files/conference/woot16/woot16-paper-grothe.pdf
* https://github.com/RUB-NDS/MS-RMS-Attacks