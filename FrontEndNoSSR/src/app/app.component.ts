import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import {ToolbarComponent} from "./util/toolbar/toolbar.component";
import {FooterComponent} from "./util/footer/footer.component";
import {BackgroundDirective} from "./background.directive";

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, ToolbarComponent, FooterComponent, BackgroundDirective],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
  title = 'Frontend';

  constructor() {
  }
}
