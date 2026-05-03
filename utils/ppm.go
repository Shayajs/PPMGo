package utils

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Width = int
type Height = int
type MaxColorValue = int
type PPMType = string

type R = uint8
type G = uint8
type B = uint8

var Ch = make(chan float32, 1) // Un canal pour communiquer l'avancement de la génération

type PPM struct {
	PPMType       PPMType
	Name          string
	Width         Width
	Height        Height
	MaxColorValue MaxColorValue
	Pixels        []Pixel
	OnGeneration  bool
}

type Pixel struct {
	R R
	G G
	B B
}

var Wg sync.WaitGroup

func (p *PPM) GeneratePPM(width Width, height Height, maxColorValue MaxColorValue) {
	// Generate a simple PPM image with a gradient
	p.PPMType = "P3"
	p.Width = width
	p.Height = height
	p.MaxColorValue = maxColorValue
	p.Pixels = make([]Pixel, 0, width*height)
	p.OnGeneration = true
}

func (p *PPM) GetSize() int {
	return len(p.Pixels)
}

func (p *PPM) GenerateRandom() {
	// 1. Initialisation de la synchronisation
	defer Wg.Done()       // Signale la fin au WaitGroup
	p.OnGeneration = true // Active le flag pour Evolve()

	// Utilisation de defer pour assurer la fermeture du canal et du flag
	defer func() {
		p.OnGeneration = false
		close(Ch) // Indispensable pour débloquer la boucle for range dans Evolve[cite: 2]
	}()

	// Initialisation du générateur aléatoire avec le temps actuel
	rgen := rand.New(rand.NewSource(time.Now().UnixNano()))

	totalPixels := float32(p.Width * p.Height)
	alreadyGeneratedPixels := 0

	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			// 2. Génération de couleurs aléatoires entre 0 et MaxColorValue
			r := uint8(rgen.Intn(p.MaxColorValue + 1))
			g := uint8(rgen.Intn(p.MaxColorValue + 1))
			b := uint8(rgen.Intn(p.MaxColorValue + 1))

			// 3. Ajout du pixel à la slice[cite: 2]
			p.AddPixel(r, g, b)

			// 4. Mise à jour de la progression[cite: 2]
			alreadyGeneratedPixels++

			// Envoi de la progression dans le canal Ch[cite: 2]
			// On ne l'envoie pas forcément à chaque pixel pour gagner en performance
			if alreadyGeneratedPixels%p.Width == 0 || alreadyGeneratedPixels == int(totalPixels) {
				Ch <- float32(alreadyGeneratedPixels) / totalPixels
			}
		}
	}
}

func (p *PPM) GenerateGradient() {
	defer Wg.Done()

	p.OnGeneration = true
	Ch <- 0.0 // On envoie une première valeur pour démarrer l'évolution
	defer func() {
		p.OnGeneration = false
		fmt.Printf("Génération de l'image terminée ! Taille de l'image : %d pixels\n", p.GetSize())
	}()

	defer close(Ch) // On ferme le canal une fois la génération terminée
	var alreadyGeneratedPixels int = 0

	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			r := uint8((x * p.MaxColorValue) / p.Width)
			g := uint8((y * p.MaxColorValue) / p.Height)
			b := uint8(p.MaxColorValue / 2)
			p.AddPixel(r, g, b)
			alreadyGeneratedPixels++
			var pixelFinished float32 = float32(alreadyGeneratedPixels)
			var totalPixels float32 = float32(p.Width) * float32(p.Height)
			Ch <- pixelFinished / totalPixels
		}
	}
}

func (p *PPM) Evolve() {

	defer Wg.Done()
	var view bool = true
	go func() {
		for p.OnGeneration {
			time.Sleep(200 * time.Millisecond)
			view = true
		}
	}()
	for ev := range Ch {
		if view {
			fmt.Printf("En cours : %.3f%% \r", ev*100)
			view = false
		}
		if !p.OnGeneration {
			break
		}
	}
	fmt.Println("\nÉvolution terminée !")
}

func (p *PPM) AddPixel(r, g, b uint8) {
	p.Pixels = append(p.Pixels, Pixel{R: R(r), G: G(g), B: B(b)})
}

func (p *PPM) SaveToFile(filename string) error {
	fmt.Printf("Sauvegarde de l'image dans le fichier %s... \n", filename)

	// 1. Création du fichier
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close() // Sécurité maximale

	// 2. On utilise un Buffer pour accumuler les données en RAM
	// avant de les envoyer au disque par gros paquets.
	writer := bufio.NewWriter(file)

	// 3. Écriture de l'en-tête (on utilise Fprintf directement sur le writer)
	fmt.Fprintf(writer, "%s\n%d %d\n%d\n", p.PPMType, p.Width, p.Height, p.MaxColorValue)

	// 4. Boucle de pixels
	for _, pixel := range p.Pixels {
		// On écrit chaque pixel directement dans le buffer
		fmt.Fprintf(writer, "%d %d %d\n", pixel.R, pixel.G, pixel.B)
	}

	// 5. TRÈS IMPORTANT : On vide le buffer dans le fichier physique
	return writer.Flush()
}

// ConvertToPPM prend un chemin de fichier (jpg/png) et renvoie une struct PPM
func ConvertToPPM(filePath string) (*PPM, error) {
	// 1. Ouvrir le fichier d'origine[cite: 1]
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 2. Décoder l'image (Go gère le format grâce aux imports '_' ci-dessus)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// 3. Initialiser notre structure PPM[cite: 1]
	ppm := &PPM{}
	ppm.GeneratePPM(width, height, 255)

	// 4. Parcourir les pixels de l'image source[cite: 1]
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// RGBA() renvoie des valeurs sur 16 bits (0-65535)
			r, g, b, _ := img.At(x, y).RGBA()

			// On convertit en 8 bits (0-255) pour notre format PPM[cite: 1]
			ppm.AddPixel(uint8(r>>8), uint8(g>>8), uint8(b>>8))
		}
	}

	return ppm, nil
}

func (p *PPM) SaveToPNG(filename string) error {
	fmt.Printf("Conversion et sauvegarde en PNG : %s... \n", filename)

	// 1. Créer une image RGBA vide avec les dimensions de ton PPM
	img := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))

	// 2. Remplir l'image avec tes pixels
	// On parcourt les pixels de ta slice p.Pixels
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			// On calcule l'index dans ta slice à plat
			idx := y*p.Width + x
			if idx < len(p.Pixels) {
				px := p.Pixels[idx]
				// On applique la couleur au pixel (x, y)
				img.Set(x, y, color.RGBA{
					R: px.R,
					G: px.G,
					B: px.B,
					A: 255, // Opacité totale
				})
			}
		}
	}

	// 3. Créer le fichier physique
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 4. Encoder l'image RGBA en format PNG vers le fichier
	return png.Encode(file, img)
}

func (p *PPM) SaveToJPG(filename string) error {
	fmt.Printf("Conversion et sauvegarde en JPG : %s... \n", filename)
	img := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			idx := y*p.Width + x
			if idx < len(p.Pixels) {
				px := p.Pixels[idx]
				img.Set(x, y, color.RGBA{R: px.R, G: px.G, B: px.B, A: 255})
			}
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	// 90 est la qualité du JPG (0-100)
	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

func (p *PPM) GeneratePerlin() {
	defer Wg.Done()
	p.OnGeneration = true
	defer func() {
		p.OnGeneration = false
		close(Ch)
	}()

	// Paramètres du bruit (ajustables pour changer l'aspect)
	scale := 0.02 // Plus c'est petit, plus les "nuages" sont grands

	// Table de permutation simplifiée pour le bruit de Perlin
	perm := make([]int, 512)
	pTable := []int{151, 160, 137, 91, 90, 15, 131, 13, 201, 95, 96, 53, 194, 233, 7, 225, 140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23, 190, 6, 148, 247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32, 57, 177, 33, 88, 237, 149, 56, 87, 174, 20, 125, 136, 171, 168, 68, 175, 74, 165, 71, 134, 139, 48, 27, 166, 77, 146, 158, 231, 83, 111, 229, 122, 60, 211, 133, 230, 220, 105, 92, 41, 55, 46, 245, 40, 244, 102, 143, 54, 65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169, 200, 196, 135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64, 52, 217, 226, 250, 124, 123, 5, 202, 38, 147, 118, 126, 255, 82, 85, 212, 207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42, 223, 183, 170, 213, 119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43, 172, 9, 129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104, 218, 246, 97, 228, 251, 34, 242, 193, 238, 210, 144, 12, 191, 179, 162, 241, 81, 51, 145, 235, 249, 14, 239, 107, 49, 192, 214, 31, 181, 199, 106, 157, 184, 84, 204, 176, 115, 121, 50, 45, 127, 4, 150, 254, 138, 236, 205, 93, 222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180}
	for i := 0; i < 256; i++ {
		perm[i] = pTable[i]
		perm[i+256] = pTable[i]
	}

	totalPixels := float32(p.Width * p.Height)
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			// Calcul du bruit pour x, y
			val := noise2D(float64(x)*scale, float64(y)*scale, perm)

			// On normalise (val est entre -1 et 1, on veut 0 à 255)
			c := uint8((val + 1) * 0.5 * float64(p.MaxColorValue))

			p.AddPixel(c, c, c) // Grisaille de Perlin

			if (y*p.Width+x)%p.Width == 0 {
				Ch <- float32(y*p.Width+x) / totalPixels
			}
		}
	}
}

// Helpers mathématiques pour le bruit de Perlin
func noise2D(x, y float64, p []int) float64 {
	X, Y := int(math.Floor(x))&255, int(math.Floor(y))&255
	xf, yf := x-math.Floor(x), y-math.Floor(y)
	u, v := fade(xf), fade(yf)
	aa, ab := p[p[X]+Y], p[p[X]+Y+1]
	ba, bb := p[p[X+1]+Y], p[p[X+1]+Y+1]

	return lerp(v, lerp(u, grad(p[aa], xf, yf), grad(p[ba], xf-1, yf)),
		lerp(u, grad(p[ab], xf, yf-1), grad(p[bb], xf-1, yf-1)))
}

func fade(t float64) float64       { return t * t * t * (t*(t*6-15) + 10) }
func lerp(t, a, b float64) float64 { return a + t*(b-a) }
func grad(hash int, x, y float64) float64 {
	switch hash & 3 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	case 3:
		return -x - y
	default:
		return 0
	}
}
