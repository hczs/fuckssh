package wizard

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func TestPasswordTestField_handleTestDoneOnlyAdvancesOnce(t *testing.T) {
	var in PasswordModeInput
	f := NewPasswordTestField(context.Background(), &in, nil, nil, nil)

	_, cmd1 := f.handleTestDone(pwTestDoneMsg{elapsed: 10 * time.Millisecond})
	if cmd1 == nil {
		t.Fatal("first success should advance to alias")
	}

	_, cmd2 := f.handleTestDone(pwTestDoneMsg{elapsed: 10 * time.Millisecond})
	if cmd2 != nil {
		t.Fatal("duplicate success should not advance again")
	}
}

func TestPasswordTestField_downAdvancesWithoutTest(t *testing.T) {
	var in PasswordModeInput
	f := NewPasswordTestField(context.Background(), &in, nil, nil, nil)
	f.textinput.SetValue("secret")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatal("down should advance without starting test")
	}
	if f.state == pwStateTesting {
		t.Fatal("down must not start connection test")
	}
}

func TestPasswordTestField_tabDoesNotAdvanceOrTest(t *testing.T) {
	var in PasswordModeInput
	f := NewPasswordTestField(context.Background(), &in, nil, nil, nil)
	f.textinput.SetValue("secret")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Fatal("tab should not advance credential field")
	}
	if f.state == pwStateTesting {
		t.Fatal("tab must not start connection test")
	}
}

func TestPasswordTestField_upGoesPrevWithoutTest(t *testing.T) {
	var in PasswordModeInput
	f := NewPasswordTestField(context.Background(), &in, nil, nil, nil)
	f.textinput.SetValue("secret")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cmd == nil {
		t.Fatal("want PrevField on up")
	}
	if f.state == pwStateTesting {
		t.Fatal("up must not start connection test")
	}
}

func TestPasswordTestField_enterStartsTest(t *testing.T) {
	var in PasswordModeInput
	f := NewPasswordTestField(context.Background(), &in, nil, nil, nil)
	f.textinput.SetValue("secret")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should start test")
	}
	if f.state != pwStateTesting {
		t.Fatalf("state = %v, want testing", f.state)
	}
}

func TestPasswordTestField_OKStateAllowsEditing(t *testing.T) {
	var in PasswordModeInput
	authOK := false
	f := NewPasswordTestField(context.Background(), &in, nil,
		func() { authOK = true },
		func() { authOK = false },
	)
	f.state = pwStateOK
	f.elapsed = 10 * time.Millisecond
	f.textinput.SetValue("secret")
	f.accessor.Set("secret")
	authOK = true

	_, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if f.state != pwStateEdit {
		t.Fatalf("state = %v, want edit after typing in OK", f.state)
	}
	if authOK {
		t.Fatal("editing after OK should clear auth test flag")
	}
	if f.textinput.Value() != "secretx" {
		t.Fatalf("value = %q, want edited password", f.textinput.Value())
	}
}

func TestPasswordTestField_enterStartsTestWhenNotLastInForm(t *testing.T) {
	var in PasswordModeInput
	f := NewPasswordTestField(context.Background(), &in, nil, nil, nil)
	f.WithPosition(huh.FieldPosition{Field: 5, LastField: 6, Group: 4, LastGroup: 5})
	f.textinput.SetValue("secret")
	if !f.keymap.Submit.Enabled() {
		t.Fatal("Submit should stay enabled when field is not last")
	}
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should start test when not last in form")
	}
}

func TestKeyIdentityField_handleTestDoneOnlyAdvancesOnce(t *testing.T) {
	var in KeyModeInput
	f := NewKeyIdentityField(context.Background(), &in, nil, nil, nil)

	_, cmd1 := f.handleTestDone(keyIDDoneMsg{elapsed: 10 * time.Millisecond})
	if cmd1 == nil {
		t.Fatal("first success should advance to alias")
	}

	_, cmd2 := f.handleTestDone(keyIDDoneMsg{elapsed: 10 * time.Millisecond})
	if cmd2 != nil {
		t.Fatal("duplicate success should not advance again")
	}
}
