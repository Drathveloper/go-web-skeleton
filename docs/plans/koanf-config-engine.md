# Plan: koanf como motor de configuración

**Fecha:** 2026-08-18
**Estado:** implementado (2026-08-18) — ver notas de cierre al final
**Alcance:** sustituir la carga de configuración actual (yaml.v3 + caarlos0/env) por
[knadh/koanf](https://github.com/knadh/koanf) como motor único, con:

1. Fichero YAML como formato principal y JSON como fallback.
2. Overrides por variables de entorno sobre el fichero.
3. Unificación de `EnvConfig` (hoy caarlos0/env con singleton global) en el mismo motor.

---

## Estado actual

- `common/config/reader_yaml.go` decodifica `config/application.yaml` desde un `embed.FS`
  con `yaml.v3` en modo estricto (`KnownFields(true)`). La estrictez es deliberada:
  protege contra la regresión documentada de `max_connection_lifetime` (clave mal escrita
  silenciosamente ignorada en producción). **Debe preservarse.**
- `common/wire/inject_configs.go` llama al reader, valida con `go-playground/validator`
  y construye `config.Store`. Ese contrato (`Store`, `model.Configuration`, validación)
  no cambia.
- `common/config/env.go` carga `model.EnvConfig` con `caarlos0/env` en un singleton
  global `config.Env` protegido por `sync.Once`, usado en `bootstrap/server.go`,
  `http/server.go`, `http/routes/routes.go` y `bootstrap/seed.go`.
- `cmd/server/main.go` embebe `config/*.yaml`; `application.yaml` no se comete,
  solo `application.example.yaml`.

## Decisiones de diseño

### Dependencias (koanf v2 es modular)

- `github.com/knadh/koanf/v2` — núcleo.
- `github.com/knadh/koanf/providers/fs` — acepta `fs.FS`; encaja con el `embed.FS`
  existente sin cambiar la arquitectura de inyección.
- `github.com/knadh/koanf/parsers/yaml` y `github.com/knadh/koanf/parsers/json`.
- `github.com/knadh/koanf/providers/env` (v2 del provider si está disponible) — overrides.
- `github.com/knadh/koanf/providers/confmap` — defaults de la capa de entorno.

### Tags del modelo

Renombrar los tags `yaml:"..."` de `common/config/model/` a `koanf:"..."`. koanf
normaliza todos los formatos a un mapa antes de deserializar, así que un tag neutro
sirve para YAML, JSON y entorno a la vez. Los tags `validate:` no cambian.
*Alternativa de coste cero si se descarta el retag:* `UnmarshalConf{Tag: "yaml"}`.

### Modo estricto (crítico)

`k.UnmarshalWithConf` con `mapstructure.DecoderConfig{ErrorUnused: true}`. Al pasar un
`DecoderConfig` propio se pierden los hooks por defecto de koanf, así que hay que
componer explícitamente:

- `StringToTimeDurationHookFunc()` — `10s`/`30m` en YAML, strings en JSON y entorno.
- `StringToSliceHookFunc(",")` — listas desde entorno (`campo1,campo2`).
- `TextUnmarshallerHookFunc()`.
- Conversión string→tipo básico para valores de entorno (siempre strings): usar
  `StringToBasicTypeHookFunc()` de mapstructure/v2 si la versión disponible lo incluye;
  en su defecto `WeaklyTypedInput: true` (trade-off: también relaja el tipado de los
  ficheros — documentarlo si se acaba usando).

El mensaje de error de clave desconocida cambia de formato (mapstructure reporta
`* '' has invalid keys: databases.postgres.pool.max_connection_lifetime`, con ruta
completa) — misma garantía, distinto texto; los tests de mensaje exacto se actualizan.

### Convención de overrides por entorno

- Prefijo `APP_`, separador de anidamiento `__` (doble guion bajo, porque las claves
  del modelo ya contienen `_`). Definirlos como constantes del paquete `config`.
- Transformación: quitar prefijo, minúsculas, `__` → `.`.
  Ejemplo: `APP_SERVER__READ_TIMEOUT=30s` → `server.read_timeout`.
- Consecuencia deliberada del modo estricto: una variable `APP_` con typo que no mapee
  a ninguna clave conocida **aborta el arranque** con un error que la nombra. Coherente
  con la filosofía del proyecto; documentarlo en README y en el example.

### Orden de capas (el último gana)

```
1. fichero  → config/application.yaml, o config/application.json si el YAML no existe
2. entorno  → env.Provider con prefijo APP_
3. unmarshal estricto + validator (sin cambios en el flujo de validación)
```

El fallback a JSON solo se activa con `fs.ErrNotExist`: un YAML malformado debe fallar,
no caer al fallback. Si no existe ninguno de los dos, el error nombra ambas rutas.

---

## Fases

### Fase 1 — Loader koanf con fallback YAML→JSON

Nuevo `common/config/reader.go` con `ReadConfig(fsys fs.FS) (*model.Configuration, error)`:

1. Intenta `config/application.yaml` (`fs.Provider` + `yaml.Parser()`).
2. Si `fs.ErrNotExist`, intenta `config/application.json` (`json.Parser()`).
3. Merge de la capa de entorno (fase 2) sobre el fichero.
4. `UnmarshalWithConf` con el modo estricto y hooks descritos arriba.
5. Errores envueltos con `constants.DefaultWrappedErrorTemplate`, como el resto del paquete.

Retag del modelo `yaml:` → `koanf:` en `common/config/model/`.

### Fase 2 — Overrides por variables de entorno

- En el mismo `ReadConfig`, tras cargar el fichero: `k.Load(env.Provider(...), nil)` con
  el prefijo `APP_` y la transformación `__` → `.`.
- Sin fichero de configuración no hay arranque (el fichero sigue siendo obligatorio);
  el entorno solo *sobreescribe*, no sustituye.

### Fase 3 — Integración wire y embed

- `common/wire/inject_configs.go`: `config.ReadYAMLConfig` → `config.ReadConfig`.
  Validación posterior sin cambios.
- `cmd/server/main.go`: glob de embed `config/*.yaml` → `config/*.yaml config/*.json`.
  **Gotcha:** un patrón `go:embed` que no casa con ningún fichero rompe la compilación,
  así que hay que cometer `config/application.example.json` (espejo exacto del example
  YAML) para que el glob JSON siempre compile. De paso documenta el formato fallback.

### Fase 4 — Unificación de EnvConfig (elimina caarlos0/env y el singleton global)

- Reescribir `common/config/env.go`: `LoadEnv(buildInfo) (*model.EnvConfig, error)`
  que **devuelve** la instancia en lugar de poblar el global `config.Env`. Internamente,
  instancia koanf separada de la del fichero:
  1. `confmap.Provider` con los defaults actuales (`gin_mode=release`, `port=8000`,
     `environment=dev`, `service_name=go-web-skeleton`, `enable_tls=false`).
     `SEED_ADMIN_*` deliberadamente sin default — preservar esa decisión y su comentario.
  2. `env.Provider` **sin prefijo pero con mapeo explícito**: la transformación solo
     traduce los nombres conocidos (`PORT`→`port`, `GIN_MODE`→`gin_mode`, `ENVIRONMENT`,
     `SERVICE_NAME`, `TLS_CERT_FILE`, `TLS_KEY_FILE`, `SEED_ADMIN_USERNAME`,
     `SEED_ADMIN_PASSWORD`, `ENABLE_TLS`) y descarta el resto del entorno del proceso.
     Sin prefijo nuevo: el contrato de despliegue actual no cambia.
  3. Unmarshal a `model.EnvConfig` (retag `env:`/`envDefault:` → `koanf:`) e inyección
     de `BuildInfo` después, como hoy.
- Threading en lugar del global:
  - `RequiredConfigs` (en `common/wire/inject_configs.go`) gana `Env *model.EnvConfig`;
    `injectConfig` llama a `LoadEnv`, de modo que `wire.Wire` recibe `buildInfo`
    (`Wire(fs, buildInfo)`).
  - `bootstrap/server.go`: elimina la llamada directa a `LoadEnv` y loguea desde
    `container.Env`.
  - `http/server.go`, `http/routes/routes.go`, `bootstrap/seed.go`: ya reciben el
    container — leer de `container.Env` en vez de `config.Env`.
- Eliminar `config.Env`, `sync.Once`, `ResetLoadEnv()` y el `//nolint:gochecknoglobals`.

### Fase 5 — Tests

Adaptar `reader_yaml_test.go` → `reader_test.go`:

- Los casos YAML existentes se conservan; actualizar aserciones de mensajes de error.
- Nuevos: solo JSON presente → carga por fallback; ambos presentes → gana YAML;
  ninguno → error con ambas rutas; YAML malformado no cae al fallback; clave
  desconocida es error duro también en JSON; duraciones parsean en ambos formatos.
- Overrides: string, duration, int, bool y slice desde entorno pisan el fichero;
  variable `APP_` desconocida → error duro. Usar `t.Setenv` (incompatible con
  `t.Parallel()` en esos tests — no marcarlos paralelos).
- Extender `TestReadYAMLConfig_ExampleConfigStillMatchesTheModel` para validar también
  `application.example.json` (un example desincronizado rompería el primer arranque de
  cada proyecto generado).
- `env_test.go`: adaptar a la firma nueva (sin global ni reset); verificar defaults,
  mapeo explícito y que variables ajenas del entorno no contaminan.

### Fase 6 — Limpieza y documentación

- Eliminar `common/config/reader_yaml.go` y `common/config/fs_reader.go` (el provider
  `fs` de koanf los reemplaza).
- `go mod tidy`: salen `gopkg.in/yaml.v3` (directa) y `github.com/caarlos0/env/v11`;
  entran los módulos koanf.
- README: documentar YAML principal + fallback JSON, la sintaxis de overrides
  (`APP_SECCION__CLAVE`), el comportamiento de fallo ante claves/variables desconocidas,
  y actualizar el paso `cp` del example mencionando ambas variantes.
- Comentar la sintaxis de override en `application.example.yaml`.

---

## Riesgos y gotchas

| Riesgo | Mitigación |
|---|---|
| Perder el modo estricto (regresión `max_connection_lifetime`) | `ErrorUnused: true` + tests de clave desconocida en YAML, JSON y entorno |
| `DecoderConfig` propio pierde hooks por defecto de koanf | Componerlos explícitamente (duration, slice, text unmarshaller, tipos básicos) |
| Glob `go:embed` sin match rompe la compilación | Cometer `application.example.json` |
| Valores de entorno son strings (int/bool fallan sin conversión) | `StringToBasicTypeHookFunc` o `WeaklyTypedInput` documentado |
| Variable `APP_` ajena en el entorno del despliegue aborta el arranque | Prefijo distintivo + documentación; es comportamiento deseado ante typos |
| Firma de `wire.Wire` cambia (gana `buildInfo`) | Cambio mecánico; compilador guía los call sites |

## Criterios de aceptación

- `make test` (o `go test ./...`) y `golangci-lint` en verde.
- Arranque con `application.yaml` idéntico al actual; arranque solo con
  `application.json` funciona; sin ninguno, falla nombrando ambas rutas.
- `APP_SERVER__READ_TIMEOUT=1s` (p. ej.) pisa el valor del fichero.
- Clave desconocida en fichero o variable `APP_` con typo → fallo de arranque que la nombra.
- No queda ninguna referencia a `config.Env` global ni a `caarlos0/env`.

## Fuera de alcance (futuro)

- Hot-reload con `file.Provider` + `Watch()` (koanf lo soporta; requeriría rediseñar
  `Store` para lecturas concurrentes).
- Providers remotos (vault, s3, etcd).
- Réplica de este esquema en el generador de scaffold si algún día emite su propio
  bootstrap en lugar de reutilizar `common/`.

---

## Notas de cierre (implementación, 2026-08-18)

Las seis fases se implementaron tal como estaban descritas. Desviaciones y
hallazgos respecto al plan:

- **No hizo falta `WeaklyTypedInput`:** go-viper/mapstructure v2.5.0 incluye
  `StringToBasicTypeHookFunc()`, así que el tipado de los ficheros sigue
  estricto. Los números de JSON (`float64`) convierten de forma nativa a
  int/int32/*int64 sin hook adicional.
- **Formato real del error de clave desconocida:**
  `decoding failed due to the following error(s):\n\n'databases.postgres.pool'
  has invalid keys: max_connection_lifetime` (ruta del struct + clave, no la
  ruta completa en la clave que anticipaba el plan). Misma garantía; los tests
  aseveran el texto real.
- **Añadido fuera de plan:** `cmd/scaffold/new.go` ahora excluye también
  `application.json` real al generar un proyecto (el `skippedFiles` solo
  cubría el YAML y habría copiado credenciales del template), y `.gitignore`
  lo cubre igualmente.
- **Revisión con ojos frescos:** sin defectos de corrección en producción.
  Dos defectos de hermeticidad en los tests nuevos, corregidos: los tests de
  `ReadConfig` y `LoadEnv` limpian del entorno del host las variables `APP_*`
  y las nueve mapeadas antes de asertar (una `PORT` o `APP_FOO` presente en
  CI rompía la suite). Nit aceptado sin cambio: `APP_SECCION__=v` (separador
  final degenerado) aborta con una clave sin nombre en el mensaje — sigue
  fallando en seguro.
- **Verificación:** `go test ./...` y `golangci-lint run` en verde; suite de
  config en verde bajo entorno hostil (`APP_FOO=1 PORT=1234 ...`); arranque
  real contra postgres/redis verificado en tres modos: normal (health OK),
  con `APP_SERVER__READ_TIMEOUT=1s` (arranca y sirve), y con
  `APP_SERVER__READ_TIMEOUTT=5s` (aborta nombrando `read_timeoutt`).
