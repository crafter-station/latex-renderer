# LaTeX Renderer

API HTTP que convierte LaTeX a HTML (con MathML) y PDF. Desplegada como Lambda container en AWS.

## Arquitectura

```
Cliente  -->  API Gateway (HTTP API)  -->  Lambda (container)  -->  latexmlc / pdflatex
                                              |
                                         ECR (imagen Docker)
```

- **$0/mes** cuando no hay trafico
- ~$0.0006 por request (10s render, 2GB RAM)
- ~$0.20/mes de ECR por la imagen

## Requisitos

- [Docker](https://docs.docker.com/get-docker/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.6
- Make
- Un perfil de AWS configurado (`aws configure --profile iamadmin-general`)

## Endpoints

### `POST /render` — LaTeX a HTML

Devuelve HTML con MathML (Presentation MathML) y CSS embebido.

```bash
curl -X POST https://TU_URL/render \
  -H "Authorization: Bearer TU_API_KEY" \
  -H "Content-Type: text/plain" \
  --data-binary @documento.tex
```

### `POST /render/pdf` — LaTeX a PDF

Devuelve el PDF compilado como binario (`application/pdf`).

```bash
curl -X POST https://TU_URL/render/pdf \
  -H "Authorization: Bearer TU_API_KEY" \
  -H "Content-Type: text/plain" \
  --data-binary @documento.tex \
  -o output.pdf
```

El body en ambos casos es el contenido `.tex` completo.

## TypeScript SDK

Disponible en [`sdk/typescript/`](sdk/typescript/).

```typescript
import { LatexRenderer } from "latex-renderer-sdk";

const client = new LatexRenderer({
  apiKey: "your-api-key",
  baseUrl: "https://your-api-url.com",
});

const html = await client.renderHTML(latexString);
const pdf = await client.renderPDF(latexString); // Uint8Array
```

Ver [SDK README](sdk/typescript/README.md) para mas detalles.

## Desarrollo local

```bash
# Construir la imagen
make build

# Levantar el container (puerto 8080, API_KEY=test123)
make run

# Probar render HTML
make test

# Probar render PDF
make test-pdf

# Ver logs
make logs

# Reiniciar despues de cambios
make restart

# Parar el container
make stop

# Limpiar todo (container + imagen)
make clean
```

Para probar manualmente:

```bash
# HTML
curl -X POST http://localhost:8080/render \
  -H "Authorization: Bearer test123" \
  -H "Content-Type: text/plain" \
  -d '\documentclass{article}\begin{document}Hello \$E=mc^2\$\end{document}'

# PDF
curl -X POST http://localhost:8080/render/pdf \
  -H "Authorization: Bearer test123" \
  -H "Content-Type: text/plain" \
  --data-binary @test.tex \
  -o output.pdf
```

## Deploy a AWS (primera vez)

```bash
# 1. Construir la imagen localmente
make build

# 2. Deploy completo (ECR + push imagen + Lambda + API Gateway)
make deploy API_KEY=tu_clave_secreta
```

Esto ejecuta todo automaticamente:
1. `terraform init`
2. Crea el repositorio ECR
3. Pushea la imagen Docker a ECR
4. Crea Lambda + API Gateway + alarma de billing

Al final muestra la URL de la API.

## Actualizar despues de cambios

Cuando modifiques `main.go`, `Dockerfile`, o cualquier archivo del proyecto:

```bash
make update
```

Esto rebuild la imagen, la pushea a ECR, y actualiza la Lambda. No toca la infra de Terraform.

## Probar en remoto

```bash
# HTML
make test-remote

# PDF
make test-remote-pdf
```

## Destruir todo (evitar gastos)

```bash
make destroy
```

Esto elimina **todos** los recursos de AWS:
- Lambda function
- API Gateway
- ECR repository (incluyendo las imagenes)
- IAM roles
- Alarmas de billing y SNS

Despues de ejecutar `make destroy`, el costo en AWS es $0.

## Tests de integracion

Requieren el container local corriendo (`make run`):

```bash
cd tests && go test -v ./...
```

## Estructura del proyecto

```
.
├── main.go                          # Servidor HTTP (Go + Gin)
├── go.mod / go.sum                  # Dependencias Go
├── Dockerfile                       # Imagen con LaTeXML + Lambda Web Adapter
├── .dockerignore
├── Makefile                         # Comandos de build, deploy, destroy
├── test.tex                         # Documento LaTeX de prueba
├── internal/
│   ├── handler/
│   │   ├── render.go                # Handler POST /render (HTML)
│   │   ├── render_pdf.go            # Handler POST /render/pdf (PDF)
│   │   └── static/css/LaTeXML.css   # CSS embebido en HTML output
│   └── middleware/
│       ├── auth.go                  # Bearer token auth
│       └── cors.go                  # CORS middleware
├── tests/
│   ├── render_pdf_test.go           # Tests de integracion
│   └── fixtures/                    # Archivos .tex para tests
├── sdk/
│   └── typescript/                  # TypeScript SDK
│       ├── src/
│       ├── package.json
│       └── README.md
└── infra/
    ├── main.tf                      # ECR + Lambda + API Gateway + Billing alarm
    ├── variables.tf                 # region, api_key, image_uri, alert_email
    ├── versions.tf                  # Provider AWS + perfil
    └── outputs.tf                   # api_url, ecr_repository_url
```

## Variables del Makefile

| Variable | Default | Descripcion |
|----------|---------|-------------|
| `API_KEY` | `test123` | Clave para el header `Authorization: Bearer` |
| `PORT` | `8080` | Puerto local |
| `PROFILE` | `iamadmin-general` | Perfil de AWS CLI |
| `REGION` | `us-east-1` | Region de AWS |

Ejemplo con valores custom:

```bash
make deploy API_KEY=mi_clave_segura REGION=eu-west-1
```
