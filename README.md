# README

## About

This is the official Wails React template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.


## Paleta

| Token de color        | Uso típico                                                 |
| --------------------- | ---------------------------------------------------------- |
| **primary**           | Color principal de la interfaz (botones destacados, links) |
| **primary-content**   | Color del texto/iconos que van sobre `primary`             |
| **secondary**         | Color secundario para diferenciar acciones o elementos     |
| **secondary-content** | Texto sobre fondo `secondary`                              |
| **accent**            | Color de acento o énfasis                                  |
| **accent-content**    | Texto sobre fondo `accent`                                 |
| **neutral**           | Fondo/elementos neutros (gris, etc.)                       |
| **neutral-content**   | Texto sobre fondo `neutral`                                |
| **base-100**          | Color de fondo principal                                   |
| **base-200**          | Fondo un poco más oscuro que `base-100`                    |
| **base-300**          | Fondo aún más oscuro, para separar secciones               |
| **base-content**      | Texto principal sobre fondos `base-*`                      |
| **info**              | Estados informativos                                       |
| **success**           | Estados de éxito                                           |
| **warning**           | Advertencias                                               |
| **error**             | Errores o estados destructivos                             |
