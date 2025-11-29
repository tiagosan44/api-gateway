Configurar autenticación vía Helm (AUTH_TYPE, JWT_SECRET, OIDC_ISSUER)

Este documento explica cómo parametrizar, de forma sencilla, las variables de autenticación del gateway cuando se despliega con el chart de Helm incluido en `deploy/helm/ai-api-gateway`.

Qué variables expone el chart hoy
- AUTH_TYPE (Values: `.config.auth.type`) → valores típicos: `jwt`, `oidc`, `both`, `mock`.
- JWT_SECRET (Secret opcional) → `.secrets.jwtSecret`. Si no se define y `AUTH_TYPE` es `jwt` o `both`, la app fallará la validación.
- OIDC_ISSUER (Values: `.config.auth.oidcIssuer`).
- OIDC_CLIENT_ID (Values: `.config.auth.oidcClientID`).
- OIDC_CLIENT_SECRET (Secret opcional) → `.secrets.oidcClientSecret`.

Ejemplo de values.yaml mínimos (dev)
```yaml
config:
  auth:
    type: "both"
    oidcIssuer: "http://mock-auth:80"
    oidcClientID: ""

secrets:
  jwtSecret: "dev-secret-change-me"
  # oidcClientSecret: ""
```

Cómo renderizar las plantillas localmente
```bash
helm template ai-api-gateway ./deploy/helm/ai-api-gateway \
  -f ./my-values.yaml
```

Instalar/actualizar en un cluster
```bash
helm upgrade --install ai-api-gateway ./deploy/helm/ai-api-gateway \
  -n default \
  -f ./my-values.yaml
```

Notas importantes
- Validación de la app: cuando `AUTH_TYPE` es `jwt` o `both`, la app exige `JWT_SECRET`. Asegúrate de definir `secrets.jwtSecret` en tus values.
- Separación de secretos: los valores sensibles se cargan en un Secret (`<release>-secrets`) y se inyectan como variables de entorno.
- Valores por defecto seguros: el chart trae defaults simples; ajusta `auth.type`, `oidcIssuer` y `jwtSecret` según tu entorno.
- Progreso incremental: este documento acompaña el avance para alinear Helm con la configuración ya soportada en Docker Compose, sin introducir aún parámetros OIDC avanzados.
