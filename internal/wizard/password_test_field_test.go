package wizard

import (
	"context"
	"testing"
	"time"
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
