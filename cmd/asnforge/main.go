package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	buildpkg "github.com/ipanalytics/ASNforge/internal/build"
	"github.com/ipanalytics/ASNforge/internal/config"
	"github.com/ipanalytics/ASNforge/internal/mmdb"
	"github.com/ipanalytics/ASNforge/internal/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "asnforge:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: asnforge <build|download|validate|inspect-ip|inspect-asn|stats|version>")
	}
	switch args[0] {
	case "build":
		fs, opts := commonFlagSet("build")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		_, err := buildpkg.Run(context.Background(), *opts)
		return err
	case "download":
		fs, opts := commonFlagSet("download")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		opts.SkipDownload = false
		_, err := buildpkg.Run(context.Background(), *opts)
		return err
	case "validate":
		fs, opts := commonFlagSet("validate")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := buildpkg.ValidateReleaseDir(opts.OutDir, opts.Strict); err != nil {
			return err
		}
		if _, _, err := mmdb.Inspect(filepath.Join(opts.OutDir, "asnforge.mmdb"), "8.8.8.8"); err != nil {
			return fmt.Errorf("MMDB cannot be opened: %w", err)
		}
		fmt.Println("PASS")
		return nil
	case "inspect-ip":
		fs, opts := commonFlagSet("inspect-ip")
		pos, err := parsePositionalCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if len(pos) != 1 {
			return fmt.Errorf("inspect-ip requires <ip>")
		}
		if opts.MMDBPath == "" {
			opts.MMDBPath = filepath.Join(opts.OutDir, "asnforge.mmdb")
		}
		rec, ok, err := mmdb.Inspect(opts.MMDBPath, pos[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no MMDB record for %s", pos[0])
		}
		return printValue(opts.Format, rec)
	case "inspect-asn":
		fs, opts := commonFlagSet("inspect-asn")
		pos, err := parsePositionalCommand(fs, args[1:])
		if err != nil {
			return err
		}
		if len(pos) != 1 {
			return fmt.Errorf("inspect-asn requires <asn>")
		}
		if opts.ASNTable == "" {
			opts.ASNTable = filepath.Join(opts.OutDir, "asnforge-asn.jsonl")
		}
		return inspectASN(opts.ASNTable, pos[0])
	case "stats":
		fs, opts := commonFlagSet("stats")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		b, err := os.ReadFile(filepath.Join(opts.OutDir, "metadata.json"))
		if err != nil {
			return err
		}
		var md buildpkg.Metadata
		if err := json.Unmarshal(b, &md); err != nil {
			return err
		}
		return printValue(opts.Format, md.Summary)
	case "version":
		fmt.Printf("asnforge %s commit=%s date=%s\n", version.Version, version.Commit, version.Date)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func parsePositionalCommand(fs *flag.FlagSet, args []string) ([]string, error) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		pos := []string{args[0]}
		if err := fs.Parse(args[1:]); err != nil {
			return nil, err
		}
		pos = append(pos, fs.Args()...)
		return pos, nil
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return fs.Args(), nil
}

func commonFlagSet(name string) (*flag.FlagSet, *config.Options) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	var opts config.Options
	config.AddCommonFlags(fs, &opts)
	return fs, &opts
}

func printValue(format string, v any) error {
	if format == "json" {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func inspectASN(path, raw string) error {
	n, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	needle := `"asn":` + strconv.FormatUint(n, 10)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, needle) {
			fmt.Println(line)
			return nil
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return fmt.Errorf("ASN %d not found in %s", n, path)
}
