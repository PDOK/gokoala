import { Component, Input } from '@angular/core'
import { NgClass } from '@angular/common'

@Component({
  selector: 'app-collection-settings',
  standalone: true,
  imports: [],
  templateUrl: './collection-settings.component.html',
  styleUrl: './collection-settings.component.css',
})
export class CollectionSettingsComponent {
  @Input() open: boolean = false

  constructor() {}
}
