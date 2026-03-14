package main

import (
	"bufio"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/otiai10/copy" // yes, i am incredibly lazy
)

var BUILD_TYPE string
var DEVTOOLS string

const DARK_MODE_CSS_PATCH = "/* ggb_patcher dark mode patch */ body { filter: invert(1) hue-rotate(180deg) brightness(1.2) contrast(0.9); }"

func pie(err error) { // pie. panic if error
	if err != nil {
		panic(err)
	}
}

func file_exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func message_box(content string) { // crude. extremely crude.

	file := path.Join(os.TempDir(), "ggb_patcher_message.vbs")
	pie(os.WriteFile(file, []byte("msgbox \""+content+"\""), 0644))

	ws := exec.Command("wscript.exe", file)
	ws.Run()

	pie(os.Remove(file))

}

func find_latest_app_version() string {

	re := regexp.MustCompile(`^app-(6\.[0-9A-Za-z\.\-]+)$`)

	local_app_data, err := os.UserCacheDir()
	pie(err)

	calc_home := path.Join(local_app_data, "GeoGebra_Calculator")

	entries, err := os.ReadDir(calc_home)
	pie(err)

	type candidate struct {
		path string
		ver  *semver.Version
	}

	var list []candidate

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		m := re.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}

		v, err := semver.NewVersion(m[1])
		if err != nil {
			continue
		}

		list = append(list, candidate{
			path: filepath.Join(calc_home, e.Name()),
			ver:  v,
		})
	}

	if len(list) == 0 {
		panic(-1)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ver.GreaterThan(list[j].ver)
	})

	return list[0].path

}

func main() {

	local_app_data, err := os.UserCacheDir()
	pie(err)

	original_squirrel_exe := path.Join(local_app_data, "GeoGebra_Calculator\\update_ggb_old.exe")
	squirrel_exe := path.Join(local_app_data, "GeoGebra_Calculator\\Update.exe")

	switch BUILD_TYPE {
	case "installer":

		current_exe, err := os.Executable()
		pie(err)

		// one file, multiple uses
		if file_exists(path.Join(filepath.Dir(current_exe), "update_ggb_old.exe")) { // act as the updater.exe file

			args_without_exe := make([]string, 0)
			for _, element := range os.Args[1:] {
				if !strings.Contains(element, "--processStart") { // usually run by the desktop/startmenu shortcut
					args_without_exe = append(args_without_exe, element)
				}
			}

			update := exec.Command(original_squirrel_exe, args_without_exe...) // run update without launching
			update.Run()

			latest := find_latest_app_version()

			// --- START CSS PATCHES ---

			fonts_path := filepath.Join(latest, "resources/app/html/css/fonts.css") // fonts.css gets loaded everywhere, so apply css patches in here

			fonts_file, err := os.Open(fonts_path)
			defer fonts_file.Close()
			pie(err)

			lines := make([]string, 0)

			scanner := bufio.NewScanner(fonts_file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			pie(scanner.Err())

			lines_new := make([]string, 0)

			for _, line := range lines {
				if !strings.Contains(line, "ggb_patcher") { // remove old patches
					lines_new = append(lines_new, line)
				}
			}

			lines_new = append(lines_new, "/*ggb_patcher*/ "+DARK_MODE_CSS_PATCH)

			pie(os.WriteFile(fonts_path, []byte(strings.Join(lines_new, "\n")), 0644))

			// --- START JS PATCHES ---

			mainjs_path := filepath.Join(latest, "resources/app/main.js")

			mainjs_file, err := os.Open(mainjs_path)
			defer mainjs_file.Close()
			pie(err)

			lines = make([]string, 0)

			scanner = bufio.NewScanner(mainjs_file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			pie(scanner.Err())

			lines_new = make([]string, 0)

			devtools_disabled := "// win.webContents.openDevTools()"
			devtools_enabled := "win.webContents.openDevTools()"

			for _, line := range lines {
				if strings.Contains(line, devtools_disabled) {
					if DEVTOOLS == "1" {
						line = strings.ReplaceAll(line, devtools_disabled, devtools_enabled)
					}
				} else if strings.Contains(line, devtools_enabled) {
					if DEVTOOLS == "0" {
						line = strings.ReplaceAll(line, devtools_enabled, devtools_disabled)
					}
				}
				lines_new = append(lines_new, line)
			}

			pie(os.WriteFile(mainjs_path, []byte(strings.Join(lines_new, "\n")), 0644))

			// --- END PATCHES ---

			launch := exec.Command(original_squirrel_exe, os.Args[1:]...) // launch with arguments
			launch.Run()

		} else { // act as the normal installer

			if !file_exists(original_squirrel_exe) { // install

				pie(os.Rename(squirrel_exe, original_squirrel_exe))
				pie(copy.Copy(current_exe, squirrel_exe))

				message_box("installed successfully")

			} else { // update

				pie(os.Remove(squirrel_exe))
				pie(copy.Copy(current_exe, squirrel_exe))

				message_box("updated successfully")

			}

		}

	case "uninstaller":

		// undo patches

		latest := find_latest_app_version()

		// --- START UNDO CSS PATCHES ---

		fonts_path := filepath.Join(latest, "resources/app/html/css/fonts.css")

		fonts_file, err := os.Open(fonts_path)
		defer fonts_file.Close()
		pie(err)

		lines := make([]string, 0)

		scanner := bufio.NewScanner(fonts_file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		pie(scanner.Err())

		lines_new := make([]string, 0)

		for _, line := range lines {
			if !strings.Contains(line, "ggb_patcher") {
				lines_new = append(lines_new, line)
			}
		}

		pie(os.WriteFile(fonts_path, []byte(strings.Join(lines_new, "\n")), 0644))

		// --- START UNDO JS PATCHES ---

		mainjs_path := filepath.Join(latest, "resources/app/main.js")

		mainjs_file, err := os.Open(mainjs_path)
		defer mainjs_file.Close()
		pie(err)

		lines = make([]string, 0)

		scanner = bufio.NewScanner(mainjs_file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		pie(scanner.Err())

		lines_new = make([]string, 0)

		devtools_disabled := "// win.webContents.openDevTools()"
		devtools_enabled := "win.webContents.openDevTools()"

		for _, line := range lines {
			if strings.Contains(line, devtools_enabled) && !strings.Contains(line, devtools_disabled) {
				line = strings.ReplaceAll(line, devtools_enabled, devtools_disabled)
			}
			lines_new = append(lines_new, line)
		}

		pie(os.WriteFile(mainjs_path, []byte(strings.Join(lines_new, "\n")), 0644))

		// --- END UNDO PATCHES ---

		// remove patcher

		if file_exists(original_squirrel_exe) {
			os.Remove(squirrel_exe)
			os.Rename(original_squirrel_exe, squirrel_exe)
		}

		message_box("uninstalled successfully")

	}
}
