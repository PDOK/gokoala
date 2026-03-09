import { inject, Pipe, PipeTransform } from '@angular/core'
import { DomSanitizer, SafeHtml } from '@angular/platform-browser'

@Pipe({
  name: 'highlight',
  standalone: true,
})
export class HighlightPipe implements PipeTransform {
  private sanitizer = inject(DomSanitizer)

  transform(properties: unknown): SafeHtml {
    const p = properties as { [key: string]: string }
    return this.sanitizer.bypassSecurityTrustHtml(p['highlight']) || ''
  }
}
