# Sheetlite

Sheetlite is a lightweight, cross-platform desktop app for viewing and editing spreadsheet documents.

## Screenshots

![Sheetlite light mode](assets/light-ss.png)

![Sheetlite dark mode](assets/dark-ss.png)

## Features

- Open spreadsheets from the file menu or by drag and drop
- Browse worksheets in a spreadsheet-style grid
- View cell formatting, merged cells, row heights, and column widths
- Edit cell values through the grid or formula bar
- Save changes back to the workbook, or use Save As
- Light, dark, and system appearance modes

## Installation

Download the latest release from [GitHub Releases](https://github.com/MegaGrindStone/sheetlite/releases/latest).

- **Windows:** download `sheetlite_windows_amd64_installer.exe` and run the installer.
- **macOS:** download `sheetlite_darwin_universal.app.zip`, unzip it, and move `sheetlite.app` to Applications.
- **Linux:** download `sheetlite_linux_amd64.tar.gz`, extract it, and run the `sheetlite` binary.

Checksums are available in `sheetlite_checksums.txt`.

macOS builds are currently unsigned and not notarized, and may remain that way. macOS may show a security warning when opening the app.

## Usage

Open a spreadsheet from the file menu or drag it into the window. Edit cells in the grid or formula bar, then save the workbook or use Save As.

## Development

Requirements:

- Go
- Node.js and pnpm
- Wails CLI

Run the app in development mode:

```sh
wails dev
```

Run tests:

```sh
go test ./...
```

Build a desktop package:

```sh
wails build
```

## Contributing

Contributions are welcome. Keep changes focused, and run tests before opening a pull request.

## Acknowledgements

Sheetlite is built with [Wails](https://wails.io/) for the cross-platform desktop shell and [Excelize](https://github.com/qax-os/excelize) for reading and writing Excel workbooks.

## License

See [LICENSE](LICENSE).
