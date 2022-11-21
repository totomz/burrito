# TODO
- [] Esponi metriche dal server (e magari fai un modulino che "wrappa" le metriche)
- [] Raccogli trace

# template-burrito
template for monorepo services

# What to do
## Import this template in your repo
Update the package reference in each `go.mod` in this repo (eg : `common/httpserver`)

## Update Auth0 settings
### Nella dashboard
- Crea una web application, copia i secrets

### Nel backend
Creap una api application, aggiungici gli scopes, copia i robini in Kargo

# Metrics
The httpserver exposer the following metrics
* number of request (labels on the status code?)