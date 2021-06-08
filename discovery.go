package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const instanceNameURI = "http://metadata.google.internal/computeMetadata/v1/instance/name"
const instanceProjectURI = "http://metadata.google.internal/computeMetadata/v1/project/project-id"
const appNameURI = "http://metadata.google.internal/computeMetadata/v1/instance/attributes/app"

const metadataKeyApp = "app"

// discoverPeers query GCP's metadata & compute instances API to obtain instances
// with the same metadata value in the key `metadataKeyApp`
func discoverPeers() (self *compute.Instance, peers []*compute.Instance) {
	client, err := google.DefaultClient(context.Background())
	if err != nil {
		log.Println("Err getting default client", err)
	}

	instanceName, err := queryMetadata(client, instanceNameURI)
	if err != nil {
		log.Println(err)
		return
	}

	projectID, err := queryMetadata(client, instanceProjectURI)
	if err != nil {
		log.Println(err)
		return
	}

	appName, err := queryMetadata(client, appNameURI)
	if err != nil {
		log.Println(err)
		return
	}

	instances, err := queryInstances(client, projectID, appName)
	if err != nil {
		log.Println(err)
		return
	}

	for _, instance := range instances {
		if len(instance.NetworkInterfaces) > 0 {
			if instance.Name == instanceName {
				self = instance
			}
			// peers including ourself
			peers = append(peers, instance)
		}
	}

	return
}

func queryMetadata(client *http.Client, uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		log.Println("Cannot parse uri", err)
		return "", err
	}

	resp, err := client.Do(&http.Request{
		Method: http.MethodGet,
		Header: http.Header{
			"Metadata-Flavor": []string{"Google"},
		},
		URL: parsed,
	})
	if err != nil {
		log.Println("Cannot query", err)
		return "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Cannot read", err)
		return "", err
	}

	return string(data), nil
}

func queryInstances(client *http.Client, projectID, appName string) ([]*compute.Instance, error) {
	computeService, err := compute.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}

	req := computeService.Instances.AggregatedList(projectID)

	result := make([]*compute.Instance, 0)

	if err := req.Pages(context.Background(), func(page *compute.InstanceAggregatedList) error {
		for _, instancesScopedList := range page.Items {
			for _, instance := range instancesScopedList.Instances {
				for _, metadata := range instance.Metadata.Items {
					// We filter only instance with the same `metadataKeyApp` value
					if metadata.Key == metadataKeyApp && *metadata.Value == appName {
						result = append(result, instance)
						break
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	return result, err
}
