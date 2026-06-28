export class AssetResolver {
  static resolve(path: string): string {
    return `./assets/${path}`
  }
}
