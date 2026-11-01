# 01 · Terminal, Git y flujo de equipo

> Cinturón blanco · 3–4 horas

## Objetivos

- Moverte con soltura en la terminal (`zsh` en macOS, `bash` en Linux).
- Usar git para lo diario: ramas, commits, diff, stash, rebase básico.
- Conocer una convención de ramas y commits típica de un equipo (`feature/<ticket>-descripcion`).
- Configurar git con SSH o HTTPS en GitHub para que Go pueda bajar módulos privados.

## 1. Terminal: lo mínimo indispensable

| Comando | Para qué |
|---|---|
| `pwd`, `ls -la`, `cd` | Dónde estoy, qué hay, moverme |
| `cat`, `less`, `head -n 20`, `tail -f archivo.log` | Leer archivos y seguir logs |
| `grep -rn "texto" src/` | Buscar texto en el código |
| `find . -name "*.go"` | Buscar archivos |
| `which go`, `go env GOPATH` | Dónde está una herramienta |
| `echo $VARIABLE`, `export VARIABLE=valor` | Variables de entorno |
| `set -a; source .env; set +a` | Cargar un `.env` completo a la sesión |
| `cmd1 \| cmd2` | Tubería: la salida de uno es la entrada del otro |
| `cmd > archivo`, `cmd 2>&1` | Redirigir salida (y errores) |
| `nc -vz -w 5 host puerto` | ¿Hay red hasta ese puerto? |
| `lsof -i :1099` | ¿Quién está usando el puerto 1099? |
| `chmod 600 archivo` | Permisos: solo el dueño lee y escribe |

**Por qué importa:** casi todos los problemas del día a día ("no conecta", "puerto ocupado", "no encuentra el módulo") se diagnostican con estas herramientas antes de tocar código.

### Variables de entorno

Un programa lee configuración del entorno (`os.Getenv` en Go). Una variable exportada en tu terminal solo vive en **esa** terminal. En el dojo usarás varias: `CONFIG` (ruta del config de `dojo-api`), `DOJO_TOKEN` (token de la API) y `GOPRIVATE`.

```bash
export GOPRIVATE=github.com/tu-org/*
go env GOPRIVATE        # confirmar
```

Para dejarla fija, agrégala a `~/.zshrc` (o `~/.bashrc`). Para `GOPRIVATE` también existe `go env -w GOPRIVATE=github.com/tu-org/*`, que la guarda en la configuración de Go.

## 2. Git para el día a día

### Configuración inicial (una sola vez)

```bash
git config --global user.name "Tu Nombre"
git config --global user.email "tu-correo@ejemplo.com"
git config --global init.defaultBranch main
git config --global pull.rebase true     # opcional: pull con rebase en vez de merge
git config --list --show-origin          # ver qué está configurado y dónde
```

### Comandos diarios

```bash
git status                       # qué cambió
git diff                         # cambios sin preparar
git diff --staged                # cambios preparados
git add archivo.go               # preparar
git commit -m "feat(almacenes): agrega filtro por activos"
git log --oneline -10            # historial corto
git switch -c feature/123-filtro-activos develop   # rama nueva desde develop
git fetch origin && git rebase origin/develop      # ponerte al día
git stash / git stash pop        # guardar cambios temporales
git check-ignore -v config.json  # ¿por qué se ignora este archivo?
```

### Convenciones de equipo (un ejemplo típico)

Cada equipo tiene las suyas; esta es una combinación muy común y es la que usa el dojo:

- **Ramas:** `feature/<ticket>-<slug>` (o `fix/...`) desde `develop` o `main`, según el flujo del equipo. El número es el ticket del gestor de tareas (Jira, GitHub Issues, Linear…).
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org): `tipo(alcance): mensaje`, con `tipo` ∈ `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `chore`. Si el equipo lo pide, el ticket va en el mensaje o en el cuerpo (`Refs: #123`).
- **PRs:** hacia la rama de integración (`develop` o `main`). Pasan revisión de un compañero y CI (tests, `go vet`, lint).
- Antes de tomar una rama como referencia, mira la fecha de su último commit: una rama abandonada no dice cómo está el sistema hoy.

### Lo que nunca va en un commit

`config.json`, `certs/`, `.env`, binarios, `docs/` generado por swag. Antes de crear un archivo con secretos, confirma que está ignorado:

```bash
git check-ignore -v config.json   # debe imprimir la regla que lo ignora
```

En `practica/dojo-api` puedes ver un ejemplo real: su [`.gitignore`](../practica/dojo-api/.gitignore) ignora `config.json` y `/docs/`, y versiona la plantilla `config.example.json`.

## 3. GitHub, SSH/HTTPS y módulos privados

En una empresa con varios microservicios es normal tener una **librería compartida interna**: un módulo Go privado con errores, config, auth y acceso a base comunes (lo verás a fondo en el módulo 09). Go, por defecto, intenta bajar todo desde `proxy.golang.org` (público) y verificarlo contra `sum.golang.org`. Para un repo privado eso falla. `GOPRIVATE` le dice a Go: "esto bájalo directo con git, sin proxy ni checksum público".

```bash
export GOPRIVATE=github.com/tu-org/*     # acepta patrones separados por coma
```

Como Go baja esos módulos **con git**, git tiene que poder autenticarse. Hay dos caminos:

### Opción A: SSH (recomendada)

```bash
ssh-keygen -t ed25519 -C "tu-correo@ejemplo.com"   # crea ~/.ssh/id_ed25519 y .pub
cat ~/.ssh/id_ed25519.pub                          # pega esto en GitHub → Settings → SSH and GPG keys
ssh -T git@github.com                              # debe saludarte con tu usuario
```

Go pide los módulos por HTTPS. Para que git use tu llave SSH en su lugar:

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### Opción B: HTTPS con token

1. En GitHub crea un *Personal Access Token* (de preferencia *fine-grained*, con solo lectura a los repos que necesitas).
2. Guárdalo en el llavero del sistema para no escribirlo cada vez: `git config --global credential.helper osxkeychain` (macOS) o el *credential manager* de tu sistema. La primera vez que git lo pida, usuario = tu usuario de GitHub y contraseña = el token.

### Prueba

```bash
git ls-remote https://github.com/tu-org/tu-libreria-privada.git   # debe listar refs, no pedir contraseña ni fallar
```

> **Ojo con la terminal de VS Code:** su integración de git define `GIT_ASKPASS` apuntando al ayudante de Electron, que a veces truena (`Node.js environment variables are disabled...`). Usa `GIT_ASKPASS= git ...` o desactiva `git.terminalAuthentication` en la configuración.

Los dos módulos del dojo (`github.com/VlV-515/dojo_backend_go/practica/...`) son públicos y no necesitan nada de esto, pero el día que tu equipo use una librería privada, este es el paso que más se atora.

## Ejercicios

1. **Terminal.** Desde la raíz del dojo, con un solo comando, cuenta cuántos archivos `.go` tiene `practica/dojo-api/src` (pista: `find` + `wc -l`).
2. **Buscar.** Encuentra con `grep` en qué archivo de `practica/dojo-api` se define `func txOrDB`.
3. **Git en seco.** Crea una carpeta temporal fuera del dojo, haz `git init`, crea `main.go`, haz dos commits con el formato Conventional Commits, crea una rama `feature/1-prueba`, cambia algo, haz `git diff` y luego `git stash` / `git stash pop`.
4. **Ignorados.** En esa carpeta crea un `.gitignore` con `config.json` y comprueba con `git check-ignore -v config.json`.
5. **GitHub.** Configura SSH (opción A) y prueba con `ssh -T git@github.com`. Si tienes acceso a algún repo privado, prueba también `git ls-remote` contra él. Si te atoras, usa `/pista`.

## Reto

Escribe en una línea un comando que muestre la fecha del último commit de cada rama remota de un repo (por ejemplo `main`, `develop` y cualquier `release/*`) para decidir cuál está viva. Pruébalo en un repo público que tengas clonado (puede ser este mismo dojo). Pista: `git for-each-ref` con `--sort` y `--format`.

## Autoevaluación

1. ¿Qué hace `GOPRIVATE` y qué pasaría sin él?
2. ¿Cuál es el formato de un commit con Conventional Commits? Escribe uno para un fix del ticket 4321.
3. ¿Por qué `git check-ignore -v` es mejor que "creo que sí está en el .gitignore"?
4. Exportaste una variable en una terminal y en otra no existe. ¿Por qué?
5. ¿Para qué sirve `url."git@github.com:".insteadOf` cuando Go baja un módulo privado?

**Listo cuando:** `ssh -T git@github.com` te reconoce y puedes hacer una rama + commit con el formato correcto sin consultar.

Siguiente: [02 · Go fundamentos](02-go-fundamentos.md)
