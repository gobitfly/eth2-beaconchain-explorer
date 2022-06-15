package utils

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/sirupsen/logrus"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// ErrInvalidSpecification indicates that a specification is of the wrong type.
var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

var gatherRegexp = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
var acronymRegexp = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")

type varInfo struct {
	Name  string
	Alt   string
	Key   string
	Field reflect.Value
	Tags  reflect.StructTag
}

// Decoder has the same semantics as Setter, but takes higher precedence.
// It is provided for historical compatibility.
type Decoder interface {
	Decode(value string) error
}

// Setter is implemented by types can self-deserialize values.
// Any type that implements flag.Value also implements Setter.
type Setter interface {
	Set(value string) error
}

func ProcessSecrets(cfg interface{}) error {
	infos, err := gatherInfo("", cfg)
	if err != nil {
		logrus.WithError(err).Error("error getting config infos")
	}

	for _, info := range infos {
		hasPrefix := strings.HasPrefix(info.Field.String(), "projects/")
		if !hasPrefix {
			continue
		}
		x, err := accessSecretVersion(info.Field.String())
		if err != nil {
			logrus.WithError(err).Error("error getting secret")
			continue
		}
		if x == nil {
			continue
		}

		field := info.Field
		typ := field.Type()

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			if field.IsNil() {
				field.Set(reflect.New(typ))
			}
			field = field.Elem()
		}

		field.SetString(*x)
	}

	return err
}

// func processField(value string, field reflect.Value) error {
// 	typ := field.Type()

// 	decoder := decoderFrom(field)
// 	if decoder != nil {
// 		return decoder.Decode(value)
// 	}
// 	// look for Set method if Decode not defined
// 	setter := setterFrom(field)
// 	if setter != nil {
// 		return setter.Set(value)
// 	}

// 	if t := textUnmarshaler(field); t != nil {
// 		return t.UnmarshalText([]byte(value))
// 	}

// 	if b := binaryUnmarshaler(field); b != nil {
// 		return b.UnmarshalBinary([]byte(value))
// 	}

// 	if typ.Kind() == reflect.Ptr {
// 		typ = typ.Elem()
// 		if field.IsNil() {
// 			field.Set(reflect.New(typ))
// 		}
// 		field = field.Elem()
// 	}

// 	switch typ.Kind() {
// 	case reflect.String:
// 		field.SetString(value)
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		var (
// 			val int64
// 			err error
// 		)
// 		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
// 			var d time.Duration
// 			d, err = time.ParseDuration(value)
// 			val = int64(d)
// 		} else {
// 			val, err = strconv.ParseInt(value, 0, typ.Bits())
// 		}
// 		if err != nil {
// 			return err
// 		}

// 		field.SetInt(val)
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		val, err := strconv.ParseUint(value, 0, typ.Bits())
// 		if err != nil {
// 			return err
// 		}
// 		field.SetUint(val)
// 	case reflect.Bool:
// 		val, err := strconv.ParseBool(value)
// 		if err != nil {
// 			return err
// 		}
// 		field.SetBool(val)
// 	case reflect.Float32, reflect.Float64:
// 		val, err := strconv.ParseFloat(value, typ.Bits())
// 		if err != nil {
// 			return err
// 		}
// 		field.SetFloat(val)
// 	case reflect.Slice:
// 		sl := reflect.MakeSlice(typ, 0, 0)
// 		if typ.Elem().Kind() == reflect.Uint8 {
// 			sl = reflect.ValueOf([]byte(value))
// 		} else if len(strings.TrimSpace(value)) != 0 {
// 			vals := strings.Split(value, ",")
// 			sl = reflect.MakeSlice(typ, len(vals), len(vals))
// 			for i, val := range vals {
// 				err := processField(val, sl.Index(i))
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 		field.Set(sl)
// 	case reflect.Map:
// 		mp := reflect.MakeMap(typ)
// 		if len(strings.TrimSpace(value)) != 0 {
// 			pairs := strings.Split(value, ",")
// 			for _, pair := range pairs {
// 				kvpair := strings.Split(pair, ":")
// 				if len(kvpair) != 2 {
// 					return fmt.Errorf("invalid map item: %q", pair)
// 				}
// 				k := reflect.New(typ.Key()).Elem()
// 				err := processField(kvpair[0], k)
// 				if err != nil {
// 					return err
// 				}
// 				v := reflect.New(typ.Elem()).Elem()
// 				err = processField(kvpair[1], v)
// 				if err != nil {
// 					return err
// 				}
// 				mp.SetMapIndex(k, v)
// 			}
// 		}
// 		field.Set(mp)
// 	}

// 	return nil
// }

func gatherInfo(prefix string, cfg interface{}) ([]varInfo, error) {
	s := reflect.ValueOf(cfg)

	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidSpecification
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidSpecification
	}
	typeOfSpec := s.Type()

	// over allocate an info array, we will extend if needed later
	infos := make([]varInfo, 0, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)
		if !f.CanSet() || isTrue(ftype.Tag.Get("ignored")) {
			continue
		}

		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					// nil pointer to a non-struct: leave it alone
					break
				}
				// nil pointer to struct: create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		// Capture information about the config variable
		info := varInfo{
			Name:  ftype.Name,
			Field: f,
			Tags:  ftype.Tag,
			Alt:   strings.ToUpper(ftype.Tag.Get("envconfig")),
		}

		// Default to the field name as the env var name (will be upcased)
		info.Key = info.Name

		// Best effort to un-pick camel casing as separate words
		if isTrue(ftype.Tag.Get("split_words")) {
			words := gatherRegexp.FindAllStringSubmatch(ftype.Name, -1)
			if len(words) > 0 {
				var name []string
				for _, words := range words {
					if m := acronymRegexp.FindStringSubmatch(words[0]); len(m) == 3 {
						name = append(name, m[1], m[2])
					} else {
						name = append(name, words[0])
					}
				}

				info.Key = strings.Join(name, "_")
			}
		}
		if info.Alt != "" {
			info.Key = info.Alt
		}
		if prefix != "" {
			info.Key = fmt.Sprintf("%s_%s", prefix, info.Key)
		}
		info.Key = strings.ToUpper(info.Key)
		infos = append(infos, info)

		if f.Kind() == reflect.Struct {
			// honor Decode if present
			if decoderFrom(f) == nil && setterFrom(f) == nil && textUnmarshaler(f) == nil && binaryUnmarshaler(f) == nil {
				innerPrefix := prefix
				if !ftype.Anonymous {
					innerPrefix = info.Key
				}

				embeddedPtr := f.Addr().Interface()
				embeddedInfos, err := gatherInfo(innerPrefix, embeddedPtr)
				if err != nil {
					return nil, err
				}
				infos = append(infos[:len(infos)-1], embeddedInfos...)

				continue
			}
		}
	}
	return infos, nil
}

// accessSecretVersion accesses the payload for the given secret version if one
// exists. The version can be a version number as a string (e.g. "5") or an
// alias (e.g. "latest").
func accessSecretVersion(name string) (*string, error) {
	// Create the client.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*45))
	defer cancel()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", err)
	}

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version: %v", err)
	}

	// WARNING: Do not print the secret in a production environment - this snippet
	// is showing how to access the secret material.
	payload := string(result.Payload.Data)
	return &payload, nil
}

func interfaceFrom(field reflect.Value, fn func(interface{}, *bool)) {
	// it may be impossible for a struct field to fail this check
	if !field.CanInterface() {
		return
	}
	var ok bool
	fn(field.Interface(), &ok)
	if !ok && field.CanAddr() {
		fn(field.Addr().Interface(), &ok)
	}
}

func decoderFrom(field reflect.Value) (d Decoder) {
	interfaceFrom(field, func(v interface{}, ok *bool) { d, *ok = v.(Decoder) })
	return d
}

func setterFrom(field reflect.Value) (s Setter) {
	interfaceFrom(field, func(v interface{}, ok *bool) { s, *ok = v.(Setter) })
	return s
}

func textUnmarshaler(field reflect.Value) (t encoding.TextUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { t, *ok = v.(encoding.TextUnmarshaler) })
	return t
}

func binaryUnmarshaler(field reflect.Value) (b encoding.BinaryUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { b, *ok = v.(encoding.BinaryUnmarshaler) })
	return b
}

func isTrue(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
