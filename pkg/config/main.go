package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

type ServerConfig struct {
	Port    int    `json:"port" env:"SERVER_PORT" default:"8080" flag:"server.port"`
	Address string `json:"address" env:"SERVER_ADDR" default:"0.0.0.0" flag:"server.addr"`
}

type Config struct {
	Server           ServerConfig `json:"server"`
	RegistryCacheTTL int          `json:"registryCacheTTL" env:"REG_CACHE_TTL" default:"300" flag:"reg-cache-ttl"`
	ArgocdNs         string       `json:"argocdNamespace" env:"ARGOCD_NS" default:"argocd" flag:"argocd-ns"`
	Kubeconfig       string       `json:"Kubeconfig" env:"KUBECONFIG" default:"" flag:"kubeconfig"`
}
func loadStruct(v reflect.Value, t reflect.Type, flagSet *flag.FlagSet, flagValues map[string]*string) error {
	for i := 0; i < t.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		// skip les champs non exportés
		if !fieldVal.CanSet() && fieldVal.Kind() != reflect.Struct {
			continue
		}

		// Si c’est une struct embarquée : appel récursif
		if fieldVal.Kind() == reflect.Struct {
			if err := loadStruct(fieldVal, fieldVal.Type(), flagSet, flagValues); err != nil {
				return err
			}
			continue
		}

		// Récupération des tags
		flagKey := fieldType.Tag.Get("flag")

		if flagKey != "" {
			ptr := flagSet.String(flagKey, "", fmt.Sprintf("override for %s", fieldType.Name))
			flagValues[fieldType.Name] = ptr
		}
	}

	return nil
}

func applyConfig(v reflect.Value, t reflect.Type, flagValues map[string]*string) error {
	for i := 0; i < t.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		if !fieldVal.CanSet() && fieldVal.Kind() != reflect.Struct {
			continue
		}

		// Struct imbriquée : récursif
		if fieldVal.Kind() == reflect.Struct {
			if err := applyConfig(fieldVal, fieldVal.Type(), flagValues); err != nil {
				return err
			}
			continue
		}

		envKey := fieldType.Tag.Get("env")
		defaultVal := fieldType.Tag.Get("default")

		flagVal := ""
		if ptr, ok := flagValues[fieldType.Name]; ok && *ptr != "" {
			flagVal = *ptr
		}

		val := ""
		switch {
		case flagVal != "":
			val = flagVal
		case envKey != "":
			if envVal, exists := os.LookupEnv(envKey); exists {
				val = envVal
			} else {
				val = defaultVal
			}
		default:
			val = defaultVal
		}

		if val == "" {
			continue
		}

		// Conversion selon type
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(val)
		case reflect.Int:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("invalid int for %s: %w", fieldType.Name, err)
			}
			fieldVal.SetInt(int64(intVal))
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("invalid bool for %s: %w", fieldType.Name, err)
			}
			fieldVal.SetBool(boolVal)
		}
	}
	return nil
}

func Load(cfg interface{}) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	flagValues := make(map[string]*string)

	if err := loadStruct(v, t, flagSet, flagValues); err != nil {
		return err
	}
	_ = flagSet.Parse(os.Args[1:])

	return applyConfig(v, t, flagValues)
}
