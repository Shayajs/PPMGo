# PPMGo

**PPMGo** est un outil en ligne de commande (CLI) écrit en Go, conçu pour manipuler le format d'image **PPM** (Portable PixMap) et servir de pont entre différents formats d'image (JPG, PNG). Il inclut des moteurs de génération de textures procédurales utilisant la puissance des **Goroutines**.

## Fonctionnalités

*   **Conversion Multi-format** : Convertissez vos fichiers entre JPG, PNG et PPM en un clin d'œil.
*   **Génération Procédurale** :
    *   **Random** : Bruit aléatoire pur (effet "neige" TV).
    *   **Gradient** : Dégradés de couleurs fluides.
    *   **Perlin Noise** : Textures organiques et nuageuses.
*   **Performance & Concurrence** : Utilisation de `sync.WaitGroup` et de `channels` pour générer les images en parallèle tout en affichant une barre de progression en temps réel.
*   **Optimisation I/O** : Écriture bufferisée via `bufio` pour gérer des images de très haute résolution (ex: 5000x5000) sans ralentissement[cite: 1, 2].

## Installation

1. Assurez-vous d'avoir [Go](https://golang.org/dl/) installé (version 1.20+ recommandée).
2. Clonez le dépôt :
   ```bash
   git clone https://github.com/votre-compte/PPMGo.git
   cd PPMGo
   ```
3. Compilez l'exécutable :
   
```bash
   go build -o ppmgo .
   ```

## Utilisation

Le programme s'utilise avec des "flags" (drapeaux) pour définir les dimensions et les formats de sortie.

### 1. Génération d'images

Générez une image de toutes pièces en spécifiant le mode (`--random`, `--gradient` ou `--perlin`) :
```bash
# Générer un bruit de Perlin HD en PNG
./ppmgo --perlin --width 1920 --height 1080 --png sortie.png

# Générer un gradient aléatoire en JPG et PPM simultanément
./ppmgo --gradient --width 800 --height 800 --jpg image.jpg --ppm image.ppm
```

### 2. Conversion d'images existantes

Passez simplement le fichier source en argument :
```bash
# Convertir un JPG en PNG
./ppmgo vacances.jpg --png vacances.png

# Convertir un PNG en PPM
./ppmgo logo.png --ppm logo.ppm
```

## Détails des Flags

| Flag | Description | Défaut |
| :--- | :--- | :--- |
| `--random` | Active le mode de bruit aléatoire | `false` |
| `--gradient`| Active le mode dégradé | `false` |
| `--perlin`  | Active le mode bruit de Perlin | `false` |
| `--width`   | Largeur de l'image générée | `800` |
| `--height`  | Hauteur de l'image générée | `600` |
| `--png`     | Chemin de sortie pour le format PNG | `""` |
| `--jpg`     | Chemin de sortie pour le format JPG | `""` |
| `--ppm`     | Chemin de sortie pour le format PPM | `""` |

## Architecture Technique

Le projet repose sur une structure modulaire :
*   **`main.go`** : Point d'entrée, gestion des arguments et orchestration des tâches.
*   **`utils/ppm.go`** :
    *   Contient la structure `PPM` et les méthodes de manipulation de pixels.
    *   **Gestion du Deadlock** : Utilisation rigoureuse de `close(channel)` et `defer Wg.Done()` pour garantir une fermeture propre des threads de calcul.
    *   **Moteur de rendu** : Utilise `image/color` et les décodeurs natifs de Go.

---

### À propos
Développé dans le cadre d'une exploration des capacités de concurrence de Go (Goroutines/Channels) et de la manipulation binaire de fichiers image.