# TODO 
- `make` controlla il cluster kube
- `make` per configurate auth0 e altri automatismi

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
3. Update the Go module path reference
```shell
shmake setup
```
