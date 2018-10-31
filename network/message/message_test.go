package message

import (
	"bytes"
	"testing"

	"github.com/gladiusio/legion/utils"
)

func BenchmarkWrite(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := New(utils.NewLegionAddress("localhost", 7946), "testmessage", []byte(`test`)).Encode()
		if i == 0 {
			b.SetBytes(int64(len(buf)))
		}
	}
}

func BenchmarkRead(b *testing.B) {
	buf := New(utils.NewLegionAddress("localhost", 7946), "testmessage", []byte(`test`)).Encode()
	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := &Message{}
		m.Decode(buf)
		// do some work to prevent cheating the benchmark:
		bytes.Equal(m.Body(), []byte(`test`))
	}
}

func TestMessageCreation(t *testing.T) {
	buf := New(utils.NewLegionAddress("localhost", 7946), "testmessage", []byte(`test`)).Encode()
	m := &Message{}
	err := m.Decode(buf)
	if err != nil {
		t.Error(err)
	}

	if m.Sender().String() != "localhost:7946" || m.Type() != "testmessage" || !bytes.Equal(m.Body(), []byte(`test`)) {
		t.Error("message not correctly decoded")
	}
}
