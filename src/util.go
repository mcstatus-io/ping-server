package main

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var (
	//go:embed icon.png
	defaultIconBytes []byte
	defaultIcon      image.Image
	//go:embed Ubuntu-B.ttf
	ubuntuBoldFontBytes []byte
	ubuntuBoldFont      font.Face
	//go:embed Ubuntu-R.ttf
	ubuntuRegularFontBytes []byte
	ubuntuRegularFont      font.Face
	//go:embed UbuntuMono-R.ttf
	ubuntuMonoFontBytes []byte
	ubuntuMonoFont      font.Face
	ubuntuMonoSmallFont font.Face
	blockedServers      *MutexArray    = nil
	ipAddressRegex      *regexp.Regexp = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

func init() {
	var err error

	if defaultIcon, err = png.Decode(bytes.NewReader(defaultIconBytes)); err != nil {
		log.Fatalf("Failed to parse default icon: %v", err)
	}

	ubuntuBold, err := truetype.Parse(ubuntuBoldFontBytes)

	if err != nil {
		log.Fatalf("Failed to parse Ubuntu Bold font: %v", err)
	}

	ubuntuBoldFont = truetype.NewFace(ubuntuBold, &truetype.Options{
		Size: 36,
	})

	ubuntuRegular, err := truetype.Parse(ubuntuRegularFontBytes)

	if err != nil {
		log.Fatalf("Failed to parse Ubuntu Regular font: %v", err)
	}

	ubuntuRegularFont = truetype.NewFace(ubuntuRegular, &truetype.Options{
		Size: 36,
	})

	ubuntuMono, err := truetype.Parse(ubuntuMonoFontBytes)

	if err != nil {
		log.Fatalf("Failed to parse Ubuntu Mono Regular font: %v", err)
	}

	ubuntuMonoFont = truetype.NewFace(ubuntuMono, &truetype.Options{
		Size: 36,
	})

	ubuntuMonoSmallFont = truetype.NewFace(ubuntuMono, &truetype.Options{
		Size: 20,
	})
}

// MutexArray is a thread-safe array for storing and checking values.
type MutexArray struct {
	List  []interface{}
	Mutex *sync.Mutex
}

// Has checks if the given value is present in the array.
func (m *MutexArray) Has(value interface{}) bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, v := range m.List {
		if v == value {
			return true
		}
	}

	return false
}

// GetBlockedServerList fetches the list of blocked servers from Mojang's session server.
func GetBlockedServerList() error {
	resp, err := http.Get("https://sessionserver.mojang.com/blockedservers")

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	// Convert []string to []interface{}
	strSlice := strings.Split(string(body), "\n")
	interfaceSlice := make([]interface{}, len(strSlice))

	for i, v := range strSlice {
		interfaceSlice[i] = v
	}

	blockedServers = &MutexArray{
		List:  interfaceSlice,
		Mutex: &sync.Mutex{},
	}

	return nil
}

// IsBlockedAddress checks if the given address is in the blocked servers list.
func IsBlockedAddress(address string) bool {
	split := strings.Split(strings.ToLower(address), ".")
	isIPAddress := ipAddressRegex.MatchString(address)

	for k := range split {
		var newAddress string

		if k == 0 {
			newAddress = strings.Join(split, ".")
		} else if isIPAddress {
			newAddress = fmt.Sprintf("%s.*", strings.Join(split[0:len(split)-k], "."))
		} else {
			newAddress = fmt.Sprintf("*.%s", strings.Join(split[k:], "."))
		}

		newAddressBytes := sha1.Sum([]byte(newAddress))
		newAddressHash := hex.EncodeToString(newAddressBytes[:])

		if blockedServers.Has(newAddressHash) {
			return true
		}
	}

	return false
}

// ParseAddress extracts the hostname and port from the given address string, and returns the default port if none is provided.
func ParseAddress(address string, defaultPort uint16) (string, uint16, error) {
	result := strings.SplitN(address, ":", 2)

	if len(result) < 1 {
		return "", 0, fmt.Errorf("'%s' does not match any known address", address)
	}

	if len(result) < 2 {
		return result[0], defaultPort, nil
	}

	port, err := strconv.ParseUint(result[1], 10, 16)

	if err != nil {
		return "", 0, err
	}

	return result[0], uint16(port), nil
}

// EncodePNG encodes an image.Image into a byte array.
func EncodePNG(img image.Image) ([]byte, error) {
	buf := &bytes.Buffer{}

	err := png.Encode(buf, img)

	return buf.Bytes(), err
}

// GetStatusIcon returns the icon of the server if it exists, or returns the default icon.
func GetStatusIcon(response *JavaStatusResponse) (image.Image, error) {
	if response == nil || !response.Online || response.Icon == nil {
		return defaultIcon, nil
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(*response.Icon, "data:image/png;base64,"))

	if err != nil {
		return nil, err
	}

	return png.Decode(bytes.NewReader(data))
}

// ScaleImageNearestNeighbor scales an image using the nearest neighbor method.
// This method is extremely inefficient, but at the small scaling values we are using the performance hit is almost immeasurable.
func ScaleImageNearestNeighbor(img image.Image, sx, sy int) image.Image {
	s := img.Bounds().Size()

	out := image.NewRGBA(image.Rect(0, 0, s.X*sx, s.Y*sy))

	for ix := 0; ix < s.X; ix++ {
		for iy := 0; iy < s.Y; iy++ {
			c := img.At(ix, iy)

			for ox := 0; ox < sx; ox++ {
				for oy := 0; oy < sy; oy++ {
					out.Set(ix*sx+ox, iy*sy+oy, c)
				}
			}
		}
	}

	return out
}

// PointerOf returns a pointer of the argument passed.
func PointerOf[T any](v T) *T {
	return &v
}
