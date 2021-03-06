package runtime

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/Southclaws/sampctl/print"
	"github.com/Southclaws/sampctl/run"
)

// NewConfigFromEnvironment creates a Config from the given environment which includes a directory
// which is searched for either `samp.json` or `samp.yaml` and environment variable versions of the
// config parameters.
func NewConfigFromEnvironment(dir string) (cfg run.Runtime, err error) {
	cfg, err = run.RuntimeFromDir(dir)
	if err != nil {
		return
	}

	// Environment variables override samp.json
	LoadEnvironmentVariables(&cfg)

	run.ApplyRuntimeDefaults(&cfg)
	cfg.ResolveRemotePlugins()

	cfg.Platform = runtime.GOOS

	err = errors.Wrap(cfg.Validate(), "runtime configuration validation failed")
	return
}

// LoadEnvironmentVariables loads Config fields from environment variables - the variable names are
// simply the `json` tag names uppercased and prefixed with `SAMP_`
// nolint:gocyclo
func LoadEnvironmentVariables(cfg *run.Runtime) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldval := v.Field(i)
		stype := t.Field(i)

		if !fieldval.CanSet() {
			continue
		}

		name := "SAMP_" + strings.ToUpper(strings.Split(t.Field(i).Tag.Get("json"), ",")[0])

		value, ok := os.LookupEnv(name)
		if !ok {
			continue
		}

		switch stype.Type.String() {
		case "*string":
			if fieldval.IsNil() {
				v := reflect.ValueOf(value)
				fieldval.Set(reflect.New(v.Type()))
			}
			fieldval.Elem().SetString(value)

		case "[]string":
			// todo: allow filterscripts via env vars
			print.Warn("cannot set filterscripts via environment variables yet")

		case "[]run.Plugin":
			// todo: plugins via env vars
			print.Warn("cannot set plugins via environment variables yet")

		case "*bool":
			valueAsBool, err := strconv.ParseBool(value)
			if err != nil {
				print.Warn("environment variable", stype.Name, "could not interpret value", value, "as boolean:", err)
				continue
			}
			if fieldval.IsNil() {
				v := reflect.ValueOf(valueAsBool)
				fieldval.Set(reflect.New(v.Type()))
			}
			fieldval.Elem().SetBool(valueAsBool)

		case "*int":
			valueAsInt, err := strconv.Atoi(value)
			if err != nil {
				print.Warn("environment variable", stype.Name, "could not interpret value", value, "as integer:", err)
				continue
			}
			if fieldval.IsNil() {
				v := reflect.ValueOf(valueAsInt)
				fieldval.Set(reflect.New(v.Type()))
			}
			fieldval.Elem().SetInt(int64(valueAsInt))

		case "*float32":
			valueAsFloat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				print.Warn("environment variable", stype.Name, "could not interpret value", value, "as float:", err)
				continue
			}
			if fieldval.IsNil() {
				v := reflect.ValueOf(valueAsFloat)
				fieldval.Set(reflect.New(v.Type()))
			}
			fieldval.Elem().SetFloat(valueAsFloat)
		default:
			panic(fmt.Sprintf("unknown kind '%s'", stype.Type.String()))
		}
	}
}
