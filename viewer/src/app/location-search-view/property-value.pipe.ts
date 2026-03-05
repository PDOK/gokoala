import { inject, Pipe, PipeTransform } from '@angular/core'
import { DomSanitizer, SafeHtml } from '@angular/platform-browser'

@Pipe({
  name: 'propertyValue',
  standalone: true,
})
export class PropertyValuePipe implements PipeTransform {
  private sanitizer = inject(DomSanitizer)

  transform(properties: unknown, key: string, isHtml: boolean): string | SafeHtml {
    const p = properties as { [key: string]: unknown }
    if (isHtml) {
      return this.sanitizer.bypassSecurityTrustHtml(p[key] as string) || ''
    }
    return (p[key] as string) || ''
  }
}
