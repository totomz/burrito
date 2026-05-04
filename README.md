# TODO 
- [] rea ha tutti i package rotti, non si genera
- [] va invocato sempre dalla root, se non trova ./services si deve incazzare
- [] deve essere llm-based?
- customization
- - [] nome del go.mod che cambia poi rea
- - [] dove salvo lo state di terraform? tutti i chart di rea sono rotti!
- - [] chiedi se mettere il nodo gcloud, firebase o aws nelle settings
- - [] chiedi anche il namespace e deploya tutto nello stesso namespace kube!
- - [] chiedimi come configurare il backend di terraform: locale, bucket s3 (che profilo aws uso?). Come chiamo il prefix dello stack?
- - [] chiedimi se devo con

## httpserver
- /_/health (e aggiorna anche le probes!)
- leva gcloud e metti dex ovunque

# template-burrito
template for monorepo services

You need [shMake](https://github.com/totomz/shmake) to use this template.

# How to use
1. Checkout an empty repository
2. Import burrito-template
```shell
git remote add template git@github.com:totomz/template-burrito.git
git pull --rebase template main
```
