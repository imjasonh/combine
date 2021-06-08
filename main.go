package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func main() {
	src1, src2, dst := os.Args[1], os.Args[2], os.Args[3]
	log.Println("combining", src1, src2, "into", dst)
	dstr, err := name.ParseReference(dst)
	if err != nil {
		log.Fatal(err)
	}

	pull := func(s string) (v1.ImageIndex, error) {
		r, err := name.ParseReference(s)
		if err != nil {
			return nil, err
		}
		return remote.Index(r)
	}
	src1i, err := pull(src1)
	if err != nil {
		log.Fatalf("pulling %q: %v", src1, err)
	}
	src2i, err := pull(src2)
	if err != nil {
		log.Fatalf("pulling %q: %v", src2, err)
	}

	plats := map[string]bool{}
	var adds []mutate.IndexAddendum
	add := func(idx v1.ImageIndex) error {
		mf, err := idx.IndexManifest()
		if err != nil {
			return err
		}
		for _, desc := range mf.Manifests {
			b, _ := json.Marshal(desc.Platform)
			if plats[string(b)] {
				return fmt.Errorf("conflicting platform %+v", *desc.Platform)
			}
			plats[string(b)] = true
			log.Printf("found platform %+v", *desc.Platform)

			img, err := idx.Image(desc.Digest)
			if err != nil {
				return err
			}
			adds = append(adds, mutate.IndexAddendum{
				Add:        img,
				Descriptor: desc,
			})
		}
		return nil
	}
	log.Println("---", src1, "---")
	if err := add(src1i); err != nil {
		log.Fatal(err)
	}
	log.Println("---", src2, "---")
	if err := add(src2i); err != nil {
		log.Fatal(err)
	}

	dsti := mutate.AppendManifests(mutate.IndexMediaType(empty.Index, types.DockerManifestList), adds...)
	mf, _ := dsti.IndexManifest()
	b, _ := json.MarshalIndent(mf, "", " ")
	log.Println(string(b))

	log.Println("pushing...")
	if err := remote.WriteIndex(dstr, dsti, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		log.Fatalf("pushing %q: %v", dst, err)
	}
	log.Println("pushed")
}
