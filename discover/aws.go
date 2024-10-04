package discover

import (
	"net"
	"net/http"
	"time"
)

const (
	awsMetadataTokenURL = "http://169.254.169.254/latest/api/token"
	awsMetadataBaseURL  = "http://169.254.169.254/latest/meta-data"
	tokenTTL            = "21600" // Token time-to-live in seconds (6 hours)
)

// NewAWSDiscoverer returns a new Amazon Web Services network discoverer
func NewAWSDiscoverer() Discoverer {
	return NewDiscoverer(
		PrivateIPv4DiscovererOption(awsPrivateIPv4),
		PublicIPv4DiscovererOption(awsPublicIPv4),
		PublicHostnameDiscovererOption(awsHostname),
	)
}

func getMetadataToken() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("PUT", awsMetadataTokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", tokenTTL)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	token, err := StandardResponseBodyString(resp)
	if err != nil {
		return "", err
	}

	return token, nil
}

func getMetadata(path string) (string, error) {
	token, err := getMetadataToken()
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", awsMetadataBaseURL+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)

	return StandardStringFromHTTP(req, client)
}

func awsPrivateIPv4() (net.IP, error) {
	ipString, err := getMetadata("/local-ipv4")
	if err != nil {
		return nil, err
	}
	return net.ParseIP(ipString), nil
}

func awsPublicIPv4() (net.IP, error) {
	ipString, err := getMetadata("/public-ipv4")
	if err != nil {
		return nil, err
	}
	return net.ParseIP(ipString), nil
}

func awsHostname() (string, error) {
	return getMetadata("/public-hostname")
}
