package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		cli.Command{
			Name: "diff",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "key-only"},
				cli.IntFlag{Name: "depth, d", Value: 1},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() != 2 {
					return cli.NewExitError("want 2 files", 0)
				}
				args := c.Args()
				errors := stats(args.Get(0), args.Get(1))
				if errors != nil {
					return cli.NewExitError(errors, 0)
				}
				f1, err := unmarshal(args.Get(0))
				if err != nil {
					return cli.NewExitError(err, 0)
				}
				f2, err := unmarshal(args.Get(1))
				if err != nil {
					return cli.NewExitError(err, 0)
				}
				fmt.Println(diff(f1, f2))
				return nil
			},
		},
		cli.Command{
			Name: "pick",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "template, t", Usage: "`FILE`"},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					return cli.NewExitError("want a file", 0)
				}
				if c.String("template") == "" {
					return cli.NewExitError("want a template", 0)
				}
				args := c.Args()
				errors := stats(args.Get(0), c.String("template"))
				if errors != nil {
					return cli.NewExitError(errors, 0)
				}
				raw, err := unmarshal(args.Get(0))
				if err != nil {
					return cli.NewExitError(err, 0)
				}
				t, err := unmarshal(c.String("template"))
				if err != nil {
					return cli.NewExitError(err, 0)
				}
				picked := pick(t, raw)
				my, err := yaml.Marshal(picked)
				fmt.Println(string(my[:]))
				return nil
			},
		},
	}

	_ = app.Run(os.Args)
}

func stats(filenames ...string) []error {
	var errs []error
	for _, filename := range filenames {
		_, err := os.Stat(filename)
		if err != nil {
			errs = append(errs, fmt.Errorf("cannot find file: %v", filename))
		}
	}
	return errs
}

func unmarshal(filename string) (interface{}, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func diff(f1, f2 interface{}) string {
	return pretty.Compare(f1, f2)
}

func pick(template, raw interface{}) interface{} {
	t := reflect.ValueOf(template)
	r := reflect.ValueOf(raw)
	switch kind := t.Kind(); kind {
	case reflect.Map:
		ret := make(map[string]interface{})
		rawOk := r.Kind() == reflect.Map
		for _, key := range t.MapKeys() {
			v := t.MapIndex(key).Interface()
			k := key.Interface().(string)
			if rawOk {
				if i := r.MapIndex(key); i.IsValid() {
					ret[k] = pick(v, i.Interface())
				} else {
					ret[k] = pick(v, nil)
				}
			} else {
				ret[k] = pick(v, nil)
			}
		}
		return ret
	case reflect.Struct:
		fmt.Println(template)
	case reflect.Array:
		fmt.Println(template)
	case reflect.Invalid:
	case reflect.Uintptr:
	case reflect.Chan:
	case reflect.Func:
	case reflect.Interface:
	case reflect.Ptr:
	case reflect.Slice:
	case reflect.UnsafePointer:
		break
	default:
		return template
	}
	return nil
}
