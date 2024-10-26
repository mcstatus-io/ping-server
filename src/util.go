package main

import (
	"crypto/rand"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	blockedServers *MutexArray[string] = nil
	hostRegEx      *regexp.Regexp      = regexp.MustCompile(`^[A-Za-z0-9-_]+(\.[A-Za-z0-9-_]+)+(:\d{1,5})?$`)
	ipAddressRegEx *regexp.Regexp      = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`)
)

// VoteOptions is the options provided as query parameters to the vote route.
type VoteOptions struct {
	IPAddress   string
	Host        string
	Port        uint16
	ServiceName string
	Username    string
	UUID        string
	Token       string
	PublicKey   string
	Timestamp   time.Time
	Timeout     time.Duration
}

// StatusOptions is the options provided as query parameters to the status route.
type StatusOptions struct {
	Query       bool
	Timeout     time.Duration
	BypassCache bool
}

// MutexArray is a thread-safe array for storing and retrieving values.
type MutexArray[T comparable] struct {
	List  []T
	Mutex *sync.Mutex
}

// Has checks if the given value is present in the array.
func (m *MutexArray[T]) Has(value T) bool {
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
		return fmt.Errorf("mojang: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	blockedServers = &MutexArray[string]{
		List:  strings.Split(string(body), "\n"),
		Mutex: &sync.Mutex{},
	}

	return nil
}

// IsBlockedAddress checks if the given address is in the blocked servers list.
func IsBlockedAddress(address string) bool {
	addressSegments := strings.Split(strings.ToLower(address), ".")
	isIPv4Address := ipAddressRegEx.MatchString(address)

	for i := range addressSegments {
		var checkAddress string

		if i == 0 {
			checkAddress = strings.Join(addressSegments, ".")
		} else if isIPv4Address {
			checkAddress = fmt.Sprintf("%s.*", strings.Join(addressSegments[0:len(addressSegments)-i], "."))
		} else {
			checkAddress = fmt.Sprintf("*.%s", strings.Join(addressSegments[i:], "."))
		}

		if blockedServers.Has(SHA256(checkAddress)) {
			return true
		}
	}

	return false
}

// ParseAddress extracts the hostname and port from the given address string, and returns the default port if none is provided.
func ParseAddress(address string, defaultPort uint16) (string, uint16, error) {
	if !hostRegEx.MatchString(address) {
		return "", 0, fmt.Errorf("'%s' does not match any known address", address)
	}

	splitHost := strings.SplitN(address, ":", 2)

	if len(splitHost) < 1 {
		return "", 0, fmt.Errorf("'%s' does not match any known address", address)
	}

	host := splitHost[0]

	if len(splitHost) < 2 {
		return host, defaultPort, nil
	}

	port, err := strconv.ParseUint(splitHost[1], 10, 16)

	if err != nil {
		return "", 0, err
	}

	return host, uint16(port), nil
}

// GetVoteOptions parses the vote options from the provided query parameters.
func GetVoteOptions(ctx *fiber.Ctx) (*VoteOptions, error) {
	result := VoteOptions{}

	// Host
	{
		result.Host = ctx.Query("host")

		if len(result.Host) < 1 {
			return nil, errors.New("missing 'host' query parameter")
		}
	}

	// Port
	{
		result.Port = uint16(ctx.QueryInt("port", 8192))
	}

	// Service Name
	{
		result.ServiceName = ctx.Query("serviceName", "mcstatus.io")

		if len(result.ServiceName) < 1 {
			return nil, fmt.Errorf("invalid 'serviceName' query parameter: %s", result.ServiceName)
		}
	}

	// Username
	{
		result.Username = ctx.Query("username")

		if len(result.Username) < 1 || len(result.Username) > 16 {
			return nil, fmt.Errorf("invalid 'username' query parameter: %s", result.Username)
		}
	}

	// UUID
	{
		result.UUID = ctx.Query("uuid")

		// TODO check for properly formatted UUID
	}

	// Public Key
	{
		result.PublicKey = ctx.Query("publickey")
	}

	// Token
	{
		result.Token = ctx.Query("token")
	}

	// IP Address
	{
		result.IPAddress = ctx.Query("ip", ctx.IP())
	}

	// Timestamp
	{
		value := ctx.Query("timestamp")

		if len(value) > 0 {
			parsedTime, err := time.Parse(time.RFC3339, value)

			if err != nil {
				return nil, fmt.Errorf("invalid 'timestamp' query parameter: %s", result.Token)
			}

			result.Timestamp = parsedTime
		} else {
			result.Timestamp = time.Now()
		}
	}

	// Timeout
	{
		result.Timeout = time.Duration(math.Max(float64(time.Second)*ctx.QueryFloat("timeout", 5.0), float64(time.Millisecond*250)))
	}

	// Test token and public key parameters
	if len(result.Token) < 1 && len(result.PublicKey) < 1 {
		return nil, errors.New("query parameter 'token', 'publickey' or both must have a value, but both were empty")
	}

	return &result, nil
}

// GetStatusOptions returns the options for status routes, with the default values filled in.
func GetStatusOptions(ctx *fiber.Ctx) (*StatusOptions, error) {
	result := &StatusOptions{}

	// Query
	{
		result.Query = ctx.QueryBool("query", true)
	}

	// Timeout
	{
		result.Timeout = time.Duration(math.Max(float64(time.Second)*ctx.QueryFloat("timeout", 5.0), float64(time.Millisecond*500)))
	}

	// Bypass Cache
	{
		result.BypassCache = Contains(config.Cache.BypassTokens, ctx.Get("Authorization"))
	}

	return result, nil
}

// GetInstanceID returns the INSTANCE_ID environment variable parsed as an unsigned 16-bit integer.
func GetInstanceID() (uint16, error) {
	if instanceID := os.Getenv("INSTANCE_ID"); len(instanceID) > 0 {
		value, err := strconv.ParseUint(instanceID, 10, 16)

		if err != nil {
			log.Fatal(err)
		}

		return uint16(value), nil
	}

	return 0, nil
}

// GetCacheKey generates a unique key used for caching status results in Redis.
func GetCacheKey(hostname string, port uint16, opts *StatusOptions) string {
	values := &url.Values{}
	values.Set("hostname", hostname)
	values.Set("port", strconv.FormatUint(uint64(port), 10))

	if opts != nil {
		values.Set("query", strconv.FormatBool(opts.Query))
	}

	return SHA256(values.Encode())
}

// Authenticate checks and requires authentication for the current request, by finding the token.
func Authenticate(ctx *fiber.Ctx) (bool, error) {
	if config.MongoDB == nil {
		return true, nil
	}

	authToken := ctx.Get("Authorization")

	if len(authToken) < 1 {
		if err := ctx.Status(http.StatusUnauthorized).SendString("Missing 'Authorization' header in request"); err != nil {
			return false, err
		}

		return false, nil
	}

	token, err := db.GetTokenByToken(authToken)

	if err != nil {
		return false, err
	}

	if token == nil {
		if err := ctx.Status(http.StatusUnauthorized).SendString("Invalid or expired authorization token, please generate another one in the dashboard"); err != nil {
			return false, err
		}

		return false, nil
	}

	if err = db.IncrementApplicationRequestCount(token.Application); err != nil {
		return false, err
	}

	if err = db.UpdateToken(
		token.ID,
		bson.M{
			"$inc": bson.M{"requestCount": 1},
			"$set": bson.M{"lastUsedAt": time.Now().UTC()},
		},
	); err != nil {
		return false, err
	}

	if err = db.UpsertRequestLog(
		bson.M{
			"application": token.Application,
			"timestamp":   GetStartOfHour(),
			"token":       token.ID,
		},
		bson.M{
			"$setOnInsert": bson.M{
				"_id": RandomHexString(16),
			},
			"$inc": bson.M{
				"requestCount": 1,
			},
		},
	); err != nil {
		return false, err
	}

	return true, nil
}

// SHA256 returns the result of hashing the input value using SHA256 algorithm.
func SHA256(input string) string {
	result := sha1.Sum([]byte(input))

	return hex.EncodeToString(result[:])
}

// PointerOf returns a pointer of the argument passed.
func PointerOf[T any](v T) *T {
	return &v
}

// Contains returns true if the array contains the value.
func Contains[T comparable](arr []T, v T) bool {
	for _, value := range arr {
		if value == v {
			return true
		}
	}

	return false
}

// Map applies the provided map function to all of the values in the array and returns the result.
func Map[I, O any](arr []I, f func(I) O) []O {
	result := make([]O, len(arr))

	for i, v := range arr {
		result[i] = f(v)
	}

	return result
}

// GetStartOfHour returns the current date and time rounded down to the start of the hour.
func GetStartOfHour() time.Time {
	return time.Now().UTC().Truncate(time.Hour)
}

// RandomHexString returns a random hexadecimal string with the specified byte length.
func RandomHexString(byteLength int) string {
	data := make([]byte, byteLength)

	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return hex.EncodeToString(data)
}
