[![Go](https://github.com/mariow/nicmanager-export/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/mariow/nicmanager-export/actions/workflows/go.yml)

# Nicmanager Export

![Screenshot](../media/screenshot-osx.png?raw=true)

## Was tut das?
Der Name sagt eigentlich schon alles: Nicmanager Export kann einen Export des Domainbestandes bei [Nicmanager](https://nicmanager.com/) ziehen. Da ich diesen Export regelmäßig brauche und immer wieder vergesse, wie er genau anzulegen ist, entstand dieses kleine Tool.
## Bedienung
Es gibt lediglich vier Eingabefelder:
1. Username: Der Benutzername bei Nicmanager (im Idealfall ein Unterbenutzer der ausschließlich Lesezugriff via API hat)
2. Passwort: Das Passwort für den obigen Benutzernamen
3. Stichtag: Es werden nur Domains exportiert die zu diesem Stichtag noch im Bestand waren, also entweder nicht oder erst nach diesem Tag gelöscht wurden.
4. Zieldatei: Name der Ausgabedatei. Die Datei wird in das Verzeichnis geschrieben in dem Nicmanager Export gestartet wurde und **es gibt viel zu wenige Absicherungen gegen versehentlichese überschreiben anderer Dateien**
Es wird eine CSV-Datei mit den Spalten *Domain*, *Order Date*, *Reg Date* und *Close Date* erstellt. 

## Warum kann das so wenig?
Der aktuelle Funktionsumfang ist exakt meine Minimalanforderung an das Tool. 

## Kompilierung
Eigentlich™ sollte sich der Code sowohl auf Linux, Mac und Windows mit "go run nicmanager-export.go" ausführen und mit "go build nicmanager-export-go" zu einem Binary kompilieren lassen. 
Getestet habe ich das bisher nur auf Linux und Mac. 

## TODO
Es fehlt noch ganz vieles, vor allem aber:
- mehr Checks für den Dateinamen, vor allem um ungültige Pfade und versehentliches überschreiben anderer Daten zu verhindern
- ein Dialog um die Zieldatei inkl. Pfad auszuwählen
- ein optionales Debug-Log
- Bedienungshinweise im Programmfenster
- mehr Optionen für die Ausgabedatei, enthaltenen Spalten etc.
- Tests
- ...

