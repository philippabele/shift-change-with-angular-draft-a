import {AfterContentInit, Component, ContentChild, Input, TemplateRef} from '@angular/core';
import {NgIf, NgTemplateOutlet} from "@angular/common";

@Component({
  selector: 'app-faq-element',
  standalone: true,
  imports: [
    NgIf,
    NgTemplateOutlet
  ],
  templateUrl: './faq-element.component.html',
  styleUrl: './faq-element.component.scss'
})
export class FaqElementComponent implements AfterContentInit {
  @ContentChild('question', { static: false, read: TemplateRef }) questionTemplateRef: TemplateRef<any> | null = null;
  @ContentChild('answer', { static: false, read: TemplateRef }) answerTemplateRef: TemplateRef<any> | null = null;
  @Input() open: boolean = false;

  ngAfterContentInit() {
  }
}
