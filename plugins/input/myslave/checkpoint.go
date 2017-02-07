package myslave

import (
	"bytes"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/siddontang/go-mysql/mysql"
)

type checkpoint struct {
	Addr     string `toml:"addr"`
	Name     string `toml:"bin_name"`
	Position uint32 `toml:"bin_pos"`

	name string

	l sync.Mutex

	lastSaveTime time.Time
}

func loadCheckpoint(name string) (*checkpoint, error) {
	var m checkpoint

	m.name = name

	f, err := os.Open(name)
	if err != nil && !os.IsNotExist(errors.Cause(err)) {
		return nil, errors.Trace(err)
	} else if os.IsNotExist(errors.Cause(err)) {
		return &m, nil
	}
	defer f.Close()

	_, err = toml.DecodeReader(f, &m)

	return &m, err
}

func (m *checkpoint) Save(force bool) error {
	m.l.Lock()
	defer m.l.Unlock()

	n := time.Now()
	if !force && n.Sub(m.lastSaveTime) < time.Second {
		return nil
	}

	var buf bytes.Buffer
	e := toml.NewEncoder(&buf)

	e.Encode(m)

	var err error
	if err = writeFileAtomic(m.name, buf.Bytes(), 0644); err != nil {
		log.Errorf("canal save master info to file %s err %v", m.name, err)
	}

	m.lastSaveTime = n

	return errors.Trace(err)
}

func (m *checkpoint) Update(name string, pos uint32) {
	m.l.Lock()
	m.Name = name
	m.Position = pos
	m.l.Unlock()
}

func (m *checkpoint) Pos() mysql.Position {
	var pos mysql.Position
	m.l.Lock()
	pos.Name = m.Name
	pos.Pos = m.Position
	m.l.Unlock()

	return pos
}

func (m *checkpoint) Close() {
	m.Save(true)
}
