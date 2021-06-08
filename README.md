# combine manifest lists

This tool combines two Docker manifest lists ("multi-arch images") into one that provides all the platforms supported by both manifest lists.

It fails if both images provide the same platforms, or if either isn't a manifest list.

# demo

```
go run ./ \
    gcr.io/distroless/static:nonroot \
    mcr.microsoft.com/windows/nanoserver:1809 \
    gcr.io/imjasonh/combined
```

This combines the [distroless](https://github.com/googlecontainertools/distroless) image providing linux platform support with an image providing Windows support.

The result is an image that provides support for both.
This image is intended to be suitable as a base image used with `ko` to provide multi-arch _and multi-OS_ support for a Go application.
