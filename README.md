# Shopcart API - Go Backend

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-v1.9-00ADD8?style=for-the-badge)
![GORM](https://img.shields.io/badge/GORM-v1.25-e3221b?style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=for-the-badge&logo=postgresql)

Ce projet est la refonte complète et optimisée de l'API Shopcart, initialement développée en PHP (Laravel), réécrite intégralement en **Go (Golang)** avec le framework **Gin**. Il offre une plateforme e-commerce robuste, performante, et conçue pour la scalabilité, gérant l'ensemble du cycle de vie des produits, paniers, commandes, paiements, livraisons et tableaux de bord d'administration.

---

## 🏗️ Architecture du Projet

L'API suit une architecture propre (Clean Architecture simplifiée) de type **MVC (Model-View-Controller)**, organisée pour garantir la maintenabilité et la séparation des responsabilités.

```text
go-backend/
├── config/           # Configuration de l'application (Variables d'environnement, DB)
├── controllers/      # Logique métier, traitement des requêtes HTTP (Endpoints)
├── middlewares/      # Intercepteurs de requêtes (Authentification JWT, Contrôle des Rôles)
├── models/           # Définition des entités de la base de données (GORM)
├── routes/           # Définition du routeur Gin centraliséisant tous les endpoints
├── uploads/          # Stockage persistant des médias (Produits, Preuves de livraison)
├── utils/            # Utilitaires réutilisables (Génération de Token JWT, Hashage Bcrypt, Slugs)
├── main.go           # Point d'entrée de l'application, instanciation du serveur
├── go.mod            # Dépendances Go
└── go.sum            # Checksums des dépendances
```

---

## 🛠️ Stack Technique

- **Langage** : Go (Golang)
- **Framework Web** : [Gin-Gonic](https://gin-gonic.com/) - Framework HTTP ultra-rapide.
- **Base de Données** : PostgreSQL
- **ORM** : [GORM](https://gorm.io/) - Pour le mapping objet-relationnel et les auto-migrations.
- **Authentification** : JSON Web Tokens (JWT) via `golang-jwt`.
- **Sécurité** : Hachage de mots de passe avec `bcrypt`. CORS configuré via `gin-contrib/cors`.
- **Gestion des configurations** : `godotenv` pour le chargement sécurisé des variables d'environnement.

---

## 👑 Fonctionnalités & Entités couvertes

Le backend couvre 100% du périmètre fonctionnel défini dans la spécification OpenAPI originale :

1. **Authentification & Utilisateurs (`User`)** : Inscription (Clients & Admins), Connexion (JWT), Gestion de profil (FCM Tokens, Statistiques par rôle). 📍 *Rôles supportés : CUSTOMER, ADMIN, VENDOR, DELIVERY, MANAGER, SUPERVISOR.*
2. **Catalogue (`Category`, `Product`, `ProductVariant`)** : CRUD complet, recherche multi-critères, liaison variante-produit (couleurs, attributs JSONB), gestion fine des stocks et téléchargement d'images.
3. **Commandes & Panier (`Cart`, `Order`, `Payment`)** : Sessions de paniers, ajout d'articles avec validation des stocks, checkout complet, intégration des paiements, historique des commandes clients.
4. **Logistique de Livraison (`DeliveryGeolocation`)** : Assignation des livreurs, upload de preuves de livraison (images sécurisées), mise à jour et suivi des statuts en temps réel.
5. **Dashboard Administratif** : Endpoints dédiés à l'extraction des KPIs, des ventes partielles par période, des produits phares et des taux de réussite de livraison.

---

## 🚀 Guide de Déploiement (Environnement local)

### Pré-requis

1. [Go (1.20+)](https://go.dev/dl/) installé.
2. [PostgreSQL](https://www.postgresql.org/download/) en cours d'exécution.

### 1. Cloner et préparer la Base de Données

Connectez-vous à votre instance PostgreSQL et créez la base de données requise :
```sql
CREATE DATABASE shopcart_db;
```

### 2. Configuration des Variables d'Environnement

À la racine du dossier `go-backend`, créez un fichier `.env` sur le modèle suivant :

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=votre_mot_de_passe
DB_NAME=shopcart_db
DB_PORT=5432

JWT_SECRET=super_secret_key_change_me_in_prod_please
```

*(Note : Si le fichier .env est absent, le système fallbackera sur ces variables par défaut ou sur les variables système).*

### 3. Installation des dépendances

Récupérez et compilez l'ensemble des modules nécessaires :

```bash
cd go-backend
go mod tidy
```

### 4. Démarrage de l'API

Exécutez l'application. Au démarrage, **GORM exécutera automatiquement les migrations (`AutoMigrate`)** pour synchroniser les modèles statiques avec le schéma PostgreSQL.

```bash
go run main.go
```

Le serveur sera alors branché et écoutera sur le port `8000` (`http://localhost:8000`).

---

## ☁️ Guide de Déploiement (Production)

Pour un déploiement professionnel (type VPS, AWS, GCP ou Render) :

### Option A : Déploiement natif
1. Compilez l'application pour l'OS cible de votre serveur :
   ```bash
   GOOS=linux GOARCH=amd64 go build -o shopcart-api main.go
   ```
2. Uploadez le binaire `shopcart-api` sur votre serveur en n'oubliant pas de créer le dossier `uploads/` à la racine de l'exécutable pour garantir la sauvegarde des fichiers persistants.
3. Lancez l'exécutable en tant que Daemon (via `systemd` ou `supervisor`).
4. Placez le serveur derrière un Reverse Proxy tel que **Nginx** pour servir en HTTPS (port 443).

### Option B : Déploiement Docker (Recommandé)

*(Note: Assurez-vous d'avoir un fichier `Dockerfile` ou docker-compose).*

1. Construisez l'image :
   ```bash
   docker build -t shopcart-go-api .
   ```
2. Démarrez l'image en incluant les volumes de persistance :
   ```bash
   docker run -d -p 8000:8000 \
     --env-file .env \
     -v $(pwd)/uploads:/app/uploads \
     shopcart-go-api
   ```

---

## 🔒 Sécurité et Permissions

Les endpoints administratifs et logistiques interrogent explicitement le rôle du porteur du jeton via nos Middlewares spécialisés :

- **`middlewares.Auth()`** : Bloque l'accès à toute requête ne présentant pas un `Bearer Token` JWT valide dans ses en-têtes. Attache sécuritairement le `user_id` et le `role` au `gin.Context`.
- **`middlewares.Management()`** : Réserve l'accès des endpoints statistiques et de gestion lourde aux acteurs système vitaux (`ADMIN`, `MANAGER`, `SUPERVISOR`).

## ✍️ Documentation API

La syntaxe attendue et le payload des réponses sont rigoureusement mappés sur la définition [OpenAPI (Swagger)](../openapi.yaml) originale du projet Laravel. L'architecture de test Postman native du projet reste 100% compatible.
