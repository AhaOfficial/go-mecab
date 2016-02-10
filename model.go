package mecab

// #include <mecab.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type Model struct {
	model *C.mecab_model_t
}

func NewModel(args map[string]string) (*Model, error) {
	// build the argument
	opts := make([]*C.char, 0, len(args)+1)
	opt := C.CString("--allocate-sentence")
	defer C.free(unsafe.Pointer(opt))
	opts = append(opts, opt)
	for k, v := range args {
		var goopt string
		if v != "" {
			goopt = fmt.Sprintf("--%s=%s", k, v)
		} else {
			goopt = "--" + k
		}
		opt := C.CString(goopt)
		defer C.free(unsafe.Pointer(opt))
		opts = append(opts, opt)
	}

	// create new MeCab model
	model := C.mecab_model_new(C.int(len(opts)), (**C.char)(&opts[0]))
	if model == nil {
		return nil, errors.New("mecab_model is not created.")
	}

	return &Model{
		model: model,
	}, nil
}

func (m *Model) Destroy() {
	C.mecab_model_destroy(m.model)
}

func (m *Model) NewMeCab() (*MeCab, error) {
	mecab := C.mecab_model_new_tagger(m.model)
	if mecab == nil {
		return nil, errors.New("mecab is not created.")
	}
	return &MeCab{mecab: mecab}, nil
}
