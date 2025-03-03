// Includes calling analyzers.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/breml/bidichk/pkg/bidichk"
	"github.com/dmitrovia/collector-metrics/internal/analaysers/mainosexit"
	"github.com/sivchari/containedctx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift" // импортируем дополнительный анализатор
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// Config — configuration file name.
const Config = `config.json`

// ConfigData describes the structure
// of the configuration file.
type ConfigData struct {
	Staticcheck []string `json:"staticcheck"`
}

func main() {
	cfg, err := getCfg()
	if err != nil {
		fmt.Println(err)

		return
	}

	multichecker.Main(
		getAnalaysers(cfg)...,
	)
}

// getCfg — get CFG.
func getCfg() (*ConfigData, error) {
	appfile, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("getCfg->Executable: %w", err)
	}

	data, err := os.ReadFile(
		filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		return nil, fmt.Errorf("getCfg->ReadFile: %w", err)
	}

	var cfg ConfigData

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// getAnalaysers — get analaysers.
func getAnalaysers(cfg *ConfigData) []*analysis.Analyzer {
	checks := []*analysis.Analyzer{
		appends.Analyzer, asmdecl.Analyzer,
		assign.Analyzer, atomic.Analyzer,
		atomicalign.Analyzer, bools.Analyzer,
		buildssa.Analyzer, buildtag.Analyzer,
		cgocall.Analyzer, composite.Analyzer,
		copylock.Analyzer, ctrlflow.Analyzer,
		deepequalerrors.Analyzer, defers.Analyzer,
		directive.Analyzer, errorsas.Analyzer,
		fieldalignment.Analyzer, findcall.Analyzer,
		framepointer.Analyzer, httpmux.Analyzer,
		httpresponse.Analyzer, ifaceassert.Analyzer,
		inspect.Analyzer, loopclosure.Analyzer,
		lostcancel.Analyzer, nilfunc.Analyzer,
		nilness.Analyzer, pkgfact.Analyzer,
		printf.Analyzer, reflectvaluecompare.Analyzer,
		shadow.Analyzer, shift.Analyzer,
		sigchanyzer.Analyzer, slog.Analyzer,
		sortslice.Analyzer, stdmethods.Analyzer,
		stdversion.Analyzer, stringintconv.Analyzer,
		structtag.Analyzer, testinggoroutine.Analyzer,
		tests.Analyzer, timeformat.Analyzer,
		unmarshal.Analyzer, unreachable.Analyzer,
		unsafeptr.Analyzer, unusedresult.Analyzer,
		unusedwrite.Analyzer, usesgenerics.Analyzer,
		waitgroup.Analyzer,
		containedctx.Analyzer,
		bidichk.NewAnalyzer(),
		mainosexit.NewCheckAnalayser(),
	}

	tmp := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		tmp[v] = true
	}

	for _, v := range staticcheck.Analyzers {
		if tmp[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	for _, v := range simple.Analyzers {
		if tmp[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	return checks
}
