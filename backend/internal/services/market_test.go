package services

import "testing"

func TestResolverTickerLocal_ConAlias(t *testing.T) {
	// YPF cotiza como YPFD.BA en Yahoo — este caso salió mal en producción una vez
	// (precio_disponible=false para todo tenedor de YPF) hasta que se agregó el alias.
	got := ResolverTickerLocal("YPF")
	if got != "YPFD" {
		t.Errorf("esperaba YPFD, obtuve %s", got)
	}
}

func TestResolverTickerLocal_SinAlias(t *testing.T) {
	got := ResolverTickerLocal("PAMP")
	if got != "PAMP" {
		t.Errorf("esperaba PAMP (sin alias, ticker igual), obtuve %s", got)
	}
}
