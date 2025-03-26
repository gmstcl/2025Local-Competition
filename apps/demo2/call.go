package main

import (
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "os"

        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/servicediscovery"
)

type Item struct {
        ID   string `json:"id"`
        Name string `json:"name"`
}

var serviceDiscovery *servicediscovery.ServiceDiscovery
var serviceName string

func init() {
        sess := session.Must(session.NewSession(&aws.Config{
                Region: aws.String(os.Getenv("REGION")),
        }))
        serviceDiscovery = servicediscovery.New(sess)
        serviceName = os.Getenv("CLOUDMAP_SERVICE_NAME")
        if serviceName == "" {
                log.Fatal("CLOUDMAP_SERVICE_NAME is required")
        }
}

func getServiceEndpoint() (string, error) {
        input := &servicediscovery.DiscoverInstancesInput{
                NamespaceName: aws.String("dev"),
                ServiceName:   aws.String(serviceName),
                MaxResults:    aws.Int64(1),
        }

        result, err := serviceDiscovery.DiscoverInstances(input)
        if err != nil || len(result.Instances) == 0 {
                return "", fmt.Errorf("failed to discover service instances: %v", err)
        }

        instance := result.Instances[0]
        return fmt.Sprintf("http://%s:%s", *instance.Attributes["AWS_INSTANCE_IPV4"], *instance.Attributes["AWS_INSTANCE_PORT"]), nil
}

func fetchItemHandler(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        if id == "" {
                http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
                return
        }

        endpoint, err := getServiceEndpoint()
        if err != nil {
                http.Error(w, fmt.Sprintf("Could not fetch service endpoint: %s", err), http.StatusInternalServerError)
                return
        }

        url := fmt.Sprintf("%s/item?id=%s", endpoint, id)
        resp, err := http.Get(url)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to fetch item: %s", err), http.StatusInternalServerError)
                return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                http.Error(w, fmt.Sprintf("Service returned status: %d", resp.StatusCode), resp.StatusCode)
                return
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to read response body: %s", err), http.StatusInternalServerError)
                return
        }

        var item Item
        err = json.Unmarshal(body, &item)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to unmarshal response body: %s", err), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(item)
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
                "status": "OK",
        })
}

func main() {
        http.HandleFunc("/fetch-item", fetchItemHandler)
        http.HandleFunc("/healthcheck", healthcheckHandler)

        port := os.Getenv("PORT")
        if port == "" {
                port = "8080"
        }
        log.Printf("Cloud Map Client running on port %s", port)
        log.Fatal(http.ListenAndServe(":"+port, nil))
}

