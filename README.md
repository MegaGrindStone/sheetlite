# Sheetlite

Sheetlite is a cross-platform, lightweight desktop app for opening spreadsheet documents without launching a full office suite.

## Features

- Open `.xlsx` files from the file menu or by drag and drop
- Browse worksheets in a spreadsheet-style grid
- View cell formatting, merged cells, row heights, and column widths
- Edit simple cell values through the grid or formula bar
- Save changes back to the workbook, or use Save As
- Light, dark, and system appearance modes

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

## Acknowledgements

Sheetlite is built with [Wails](https://wails.io/) for the cross-platform desktop shell and [Excelize](https://github.com/qax-os/excelize) for reading and writing Excel workbooks.
