
import { BrowserModule } from '@angular/platform-browser';
import { AppComponent } from './app.component';
import { createCustomElement } from '@angular/elements';
import { ObjectInfoComponent } from './object-info/object-info.component';
import { NgModule, Injector } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';


@NgModule({
  declarations: []
  ,
  providers: [],
  bootstrap: [],
  imports: [
    BrowserModule,
    HttpClientModule,
  
    AppComponent
  ], 
 

})

export class AppModule {
  constructor(private injector: Injector) {
    const webComponent = createCustomElement(AppComponent, { injector });
    customElements.define('app-vectortile-view', webComponent);
    const webObjectInfo = createCustomElement(ObjectInfoComponent, { injector });
    customElements.define('app-objectinfo-view', webObjectInfo);
  }
  ngDoBootstrap() { }
}
