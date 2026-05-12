package autorunner

//func TestNewRunnerRequiresMissingWine(t *testing.T) {
//	_, err := newRunner("/tmp/prefix", DependencyStatus{})
//	if err == nil {
//		t.Fatal("newRunner error = nil, want error")
//	}
//}
//
//func TestNewRunnerUsesWineWhenWineInstalled(t *testing.T) {
//	r, err := newRunner("/tmp/prefix", DependencyStatus{
//		WinePath: "/usr/bin/wine",
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	wineRunner, ok := r.(*wine.Runner)
//	if !ok {
//		t.Fatalf("runner type = %T, want *wine.Runner", r)
//	}
//	if got := wineRunner.DefaultPrefix; got != "/tmp/prefix" {
//		t.Fatalf("DefaultPrefix = %q, want %q", got, "/tmp/prefix")
//	}
//}
