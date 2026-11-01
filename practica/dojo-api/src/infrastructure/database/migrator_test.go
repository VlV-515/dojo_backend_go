package database

import "testing"

func TestCargarMigracionesOrdenadas(t *testing.T) {
	migraciones, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	if len(migraciones) == 0 {
		t.Fatal("no hay migraciones embebidas")
	}
	for i := 1; i < len(migraciones); i++ {
		if migraciones[i].Version <= migraciones[i-1].Version {
			t.Errorf("versiones fuera de orden o repetidas: %d después de %d", migraciones[i].Version, migraciones[i-1].Version)
		}
	}
}
