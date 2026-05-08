package wine

import (
	"testing"

	"github.com/DarlingGoose/gr"
)

func TestList(t *testing.T) {

	_, err := List(t.Context(), "wine", []string{}, gr.WithWinePrefix("/home/n9s/.local/vntext/ksh-dlexe"), gr.WithName("KSH_dl.exe"))
	if err != nil {
		t.Fatal(err)
	}
}
