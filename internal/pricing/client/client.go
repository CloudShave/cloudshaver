package client

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"
)

const (
    DefaultPricingRegion = "us-east-1"
    BaseURL             = "https://pricing.%s.amazonaws.com/offers/v1.0/aws"
    IndexFile           = "index.json"
    CacheExpiration     = 24 * time.Hour
)

// PricingClient handles AWS pricing API interactions
type PricingClient struct {
    httpClient  *http.Client
    region      string
    cache       map[string]map[string]*CachedResponse
    cacheMutex  sync.RWMutex
}

type CachedResponse struct {
    Data      []byte
    Timestamp time.Time
}

type ServiceIndex struct {
    FormatVersion   string    `json:"formatVersion"`
    Disclaimer      string    `json:"disclaimer"`
    PublicationDate time.Time `json:"publicationDate"`
    Offers          map[string]struct {
        CurrentVersion      string            `json:"currentVersion"`
        CurrentRegionIndex string            `json:"currentRegionIndexUrl"`
        Regions            map[string]string `json:"regions"`
    } `json:"offers"`
}

// NewPricingClient creates a new AWS pricing API client
func NewPricingClient(region string) *PricingClient {
    if region == "" {
        region = DefaultPricingRegion
    }

    return &PricingClient{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        region: region,
        cache:  make(map[string]map[string]*CachedResponse),
    }
}

// GetServiceIndex retrieves the main AWS pricing index
func (c *PricingClient) GetServiceIndex() (*ServiceIndex, error) {
    url := fmt.Sprintf("%s/%s", c.getBaseURL(), IndexFile)
    data, err := c.fetchWithCache(url, c.region, "")
    if err != nil {
        return nil, err
    }

    var index ServiceIndex
    if err := json.Unmarshal(data, &index); err != nil {
        return nil, fmt.Errorf("failed to parse service index: %v", err)
    }

    return &index, nil
}

// GetServicePricing retrieves pricing data for a specific service
func (c *PricingClient) GetServicePricing(service, region string) ([]byte, error) {
    index, err := c.GetServiceIndex()
    if err != nil {
        return nil, err
    }

    offer, exists := index.Offers[service]
    if !exists {
        return nil, fmt.Errorf("service %s not found in pricing index", service)
    }

    var pricingURL string
    if region != "" {
        regionURL, exists := offer.Regions[region]
        if !exists {
            return nil, fmt.Errorf("region %s not found for service %s", region, service)
        }
        pricingURL = regionURL
    } else {
        pricingURL = offer.CurrentRegionIndex
    }

    // Convert relative URL to absolute
    if pricingURL[0] == '/' {
        pricingURL = c.getBaseURL() + pricingURL
    }

    return c.fetchWithCache(pricingURL, region, service)
}

func (c *PricingClient) fetchWithCache(url, region, service string) ([]byte, error) {
    c.cacheMutex.RLock()
    if regionCache, ok := c.cache[region]; ok {
        if cached, ok := regionCache[service]; ok {
            if time.Since(cached.Timestamp) < CacheExpiration {
                c.cacheMutex.RUnlock()
                return cached.Data, nil
            }
        }
    }
    c.cacheMutex.RUnlock()

    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch pricing data: %w", err)
    }
    defer resp.Body.Close()

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    c.cacheMutex.Lock()
    defer c.cacheMutex.Unlock()

    if _, ok := c.cache[region]; !ok {
        c.cache[region] = make(map[string]*CachedResponse)
    }
    c.cache[region][service] = &CachedResponse{
        Data:      data,
        Timestamp: time.Now(),
    }

    return data, nil
}

func (c *PricingClient) getBaseURL() string {
    return fmt.Sprintf(BaseURL, c.region)
}

// ClearCache clears the pricing data cache
func (c *PricingClient) ClearCache() {
    c.cacheMutex.Lock()
    c.cache = make(map[string]map[string]*CachedResponse)
    c.cacheMutex.Unlock()
}
