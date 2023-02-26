package internal

import (
    "io"
    "log"
    "os"
    "path/filepath"

    "github.com/JouleJ/socnet/core"
)

type resource struct {
    name string
    data []byte
}

func (r *resource) Name() string {
    return r.name
}

func (r *resource) Content() []byte {
    if r.data != nil {
        return r.data
    }

    f, err := os.Open(r.name)
    defer f.Close()

    if err != nil {
        log.Printf("Failed to open file %v due to %v\n", r.name, err)
        return r.data
    }

    bytes, err := io.ReadAll(f)
    if err != nil {
        log.Printf("Failed to read file %v due to %v\n", r.name, err)
        return r.data
    }

    r.data = bytes
    return r.data
}

func NewResource(name string) core.Resource {
    return &resource{name: name}
}

type resourceManager struct {
    resources []core.Resource
}

func walkResourceFolder(resources *[]core.Resource, path string) {
    log.Printf("walkResourceFolder: path=%v\n", path)

    entries, err := os.ReadDir(path)
    if err != nil {
        log.Printf("walkResourceFolder: failed due to %v\n", err)
        return
    }

    for _, entry := range entries {
        entryPath := filepath.Join(path, entry.Name())
        if entry.IsDir() {
            walkResourceFolder(resources, entryPath)
        } else {
            *resources = append(*resources, NewResource(entryPath))
        }
    }
}

func (rm *resourceManager) GetList() []core.Resource {
    if rm.resources != nil {
        return rm.resources
    }

    resourcePath := os.Getenv("RESOURCE_PATH")
    log.Printf("RESOURCE_PATH=%v\n", resourcePath)
    if resourcePath == "" {
        log.Fatalf("RESOURCE_PATH is empty!\n")
    }

    rm.resources = []core.Resource{}
    walkResourceFolder(&rm.resources, resourcePath)

    return rm.resources
}

func NewResourceManager() core.ResourceManager {
    return &resourceManager{}
}
