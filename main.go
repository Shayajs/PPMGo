package main

import (
	"flag"
	"fmt"
	"ppmgo/utils"
)

func main() {
	// --- 1. Définition de TOUS les drapeaux au début ---
	genRandom := flag.Bool("random", false, "Génère une image de bruit aléatoire")
	genGradient := flag.Bool("gradient", false, "Génère un gradient de couleurs")
	genPerlin := flag.Bool("perlin", false, "Génère du bruit de Perlin (organique)")

	width := flag.Int("width", 800, "Largeur de l'image à générer")
	height := flag.Int("height", 600, "Hauteur de l'image à générer")

	toPNG := flag.String("png", "", "Chemin de sortie pour un fichier PNG")
	toPPM := flag.String("ppm", "", "Chemin de sortie pour un fichier PPM")
	toJPG := flag.String("jpg", "", "Chemin de sortie pour un fichier JPG")

	// Analyse des arguments
	flag.Parse()

	var ppm *utils.PPM

	// --- 2. Logique de sélection : Génération ou Conversion ---
	if *genRandom || *genGradient || *genPerlin {
		// MODE GÉNÉRATION
		fmt.Printf("Démarrage de la génération (%dx%d)... \n", *width, *height)

		ppm = &utils.PPM{}
		ppm.GeneratePPM(*width, *height, 255) // Initialisation de la structure

		// Préparation de la synchronisation pour les Goroutines
		utils.Wg.Add(2)

		// Lancement de l'affichage de la progression
		go ppm.Evolve()

		// Sélection du moteur de génération
		if *genRandom {
			go ppm.GenerateRandom()
		} else if *genPerlin {
			go ppm.GeneratePerlin() // Ta nouvelle fonction !
		} else {
			go ppm.GenerateGradient()
		}

		// On attend que le travail soit fini
		utils.Wg.Wait()

	} else {
		// MODE CONVERSION
		args := flag.Args()
		if len(args) < 1 {
			fmt.Println("Usage :")
			fmt.Println("  Génération : ppmgo --perlin --width 1024 --png out.png")
			fmt.Println("  Conversion : ppmgo source.jpg --png out.png")
			return
		}

		sourcePath := args[0]
		var err error
		ppm, err = utils.ConvertToPPM(sourcePath)
		if err != nil {
			fmt.Printf("Erreur lors de la lecture de %s : %v\n", sourcePath, err)
			return
		}
	}

	// --- 3. Sauvegarde multi-format ---
	if ppm != nil {
		if *toPNG != "" {
			if err := ppm.SaveToPNG(*toPNG); err == nil {
				fmt.Println("✓ Image sauvegardée en PNG")
			}
		}

		if *toPPM != "" {
			if err := ppm.SaveToFile(*toPPM); err == nil {
				fmt.Println("✓ Image sauvegardée en PPM")
			}
		}

		if *toJPG != "" {
			if err := ppm.SaveToJPG(*toJPG); err == nil {
				fmt.Println("✓ Image sauvegardée en JPG")
			}
		}
	} else {
		fmt.Println("Erreur : Aucune donnée d'image à sauvegarder.")
	}
}
