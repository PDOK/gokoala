import { NgModule, Injector } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import {AppComponent } from './app.component';
import  { createCustomElement } from '@angular/elements';


@NgModule({
  declarations: [
    AppComponent,

  ],
  imports: [
    BrowserModule
    
   
  ],
  providers: [
    
  ],
  bootstrap: []
})

export class AppModule {
  constructor(private injector: Injector) {
    const webComponent = createCustomElement(AppComponent, {injector});
    customElements.define('app-vectortile-view', webComponent);
  }
  ngDoBootstrap(){}
}
