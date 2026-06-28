import Phaser from "phaser"

export class MenuIntroScene extends Phaser.Scene {
  private slides = ["menu-prev-1", "menu-prev-2"]
  private currentIndex = 0
  constructor() {
    super("menu-intro")
  }

  create(): void {
    this.showSlide()
  }

  private showSlide(): void {
    const key = this.slides[this.currentIndex]
    const image = this.add.image(
      this.scale.width / 2,
      this.scale.height / 2,
      key,
    )

    image.setDisplaySize(this.scale.width, this.scale.height)
    image.setAlpha(0)

    this.tweens.add({
      targets: image,
      alpha: 1,
      duration: 500,

      onComplete: () => {
        this.time.delayedCall(1200, () => {
          this.tweens.add({
            targets: image,
            alpha: 0,
            duration: 500,
            onComplete: () => {
              image.destroy()
              this.currentIndex++
              if (this.currentIndex >= this.slides.length) {
                this.scene.start("main-menu")
                return
              }
              this.showSlide()
            },
          })
        })
      },
    })
  }
}
