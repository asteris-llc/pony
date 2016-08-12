package tf

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
)

type UiHook struct {
	terraform.NilHook

	Ui cli.Ui

	l         sync.Mutex
	once      sync.Once
	resources map[string]uiResourceOp
	ui        cli.Ui
}

type uiResourceOp byte

const (
	uiResourceUnknown uiResourceOp = iota
	uiResourceCreate
	uiResourceModify
	uiResourceDestroy
)

func NewUiHook(ui cli.Ui) *UiHook {
	return &UiHook{
		Ui: ui,
	}
}

func getName(d *terraform.InstanceDiff, id string) string {
	if rname, ok := d.Attributes["name"]; ok {
		return rname.New
	} 

	return id
}

func (h *UiHook) PreApply(
	n *terraform.InstanceInfo,
	s *terraform.InstanceState,
	d *terraform.InstanceDiff) (terraform.HookAction, error) {
	h.once.Do(h.init)

	id := n.HumanId()

	op := uiResourceModify
	if d.Destroy {
		op = uiResourceDestroy
	} else if s.ID == "" {
		op = uiResourceCreate
	}

	h.l.Lock()
	h.resources[id] = op
	h.l.Unlock()

	var operation string
	switch op {
	case uiResourceModify:
		operation = "Modifying..."
	case uiResourceDestroy:
		operation = "Destroying..."
	case uiResourceCreate:
		operation = "Creating..."
	case uiResourceUnknown:
		return terraform.HookActionContinue, nil
	}


	h.ui.Output(fmt.Sprintf(
		"%s: %s",
		getName(d, id),
		operation))

	return terraform.HookActionContinue, nil
}

func (h *UiHook) PostApply(
	n *terraform.InstanceInfo,
	s *terraform.InstanceState,
	applyerr error) (terraform.HookAction, error) {
	id := n.HumanId()

	h.l.Lock()
	op := h.resources[id]
	delete(h.resources, id)
	h.l.Unlock()

	var msg string
	switch op {
	case uiResourceModify:
		msg = "Modifications complete"
	case uiResourceDestroy:
		msg = "Destruction complete"
	case uiResourceCreate:
		msg = "Creation complete"
	case uiResourceUnknown:
		return terraform.HookActionContinue, nil
	}

	if applyerr != nil {
		// Errors are collected and printed in ApplyCommand, no need to duplicate
		return terraform.HookActionContinue, nil
	}

	h.ui.Output(fmt.Sprintf(
		"%s: %s",
		id,
		msg))

	return terraform.HookActionContinue, nil
}

func (h *UiHook) PreDiff(
	n *terraform.InstanceInfo,
	s *terraform.InstanceState) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (h *UiHook) PreProvision(
	n *terraform.InstanceInfo,
	provId string) (terraform.HookAction, error) {
	id := n.HumanId()
	h.ui.Output(fmt.Sprintf(
		"%s: Provisioning with '%s'...",
		id, provId))
	return terraform.HookActionContinue, nil
}

func (h *UiHook) ProvisionOutput(
	n *terraform.InstanceInfo,
	provId string,
	msg string) {
	id := n.HumanId()
	var buf bytes.Buffer

	prefix := fmt.Sprintf("%s (%s): ", id, provId)
	s := bufio.NewScanner(strings.NewReader(msg))
	s.Split(scanLines)
	for s.Scan() {
		line := strings.TrimRightFunc(s.Text(), unicode.IsSpace)
		if line != "" {
			buf.WriteString(fmt.Sprintf("%s%s\n", prefix, line))
		}
	}

	h.ui.Output(strings.TrimSpace(buf.String()))
}

func (h *UiHook) PreRefresh(
	n *terraform.InstanceInfo,
	s *terraform.InstanceState) (terraform.HookAction, error) {
	h.once.Do(h.init)

	id := n.HumanId()
	h.ui.Output(fmt.Sprintf(
		"%s: Refreshing state... (ID: %s)",
		id,
		s.ID))
	return terraform.HookActionContinue, nil
}

func (h *UiHook) init() {
	h.resources = make(map[string]uiResourceOp)

	// Wrap the ui so that it is safe for concurrency regardless of the
	// underlying reader/writer that is in place.
	h.ui = &cli.ConcurrentUi{Ui: h.Ui}
}

// scanLines is basically copied from the Go standard library except
// we've modified it to also fine `\r`.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
