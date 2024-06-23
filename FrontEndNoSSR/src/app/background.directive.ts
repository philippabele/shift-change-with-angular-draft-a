import { Directive, ElementRef, Renderer2, OnInit, OnDestroy, Inject } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { Subscription } from 'rxjs';
import { DOCUMENT } from "@angular/common";
import {HttpClient} from "@angular/common/http";

@Directive({
  selector: '[appBackground]',
  standalone: true
})
export class BackgroundDirective implements OnInit, OnDestroy {
  private routerSubscription: Subscription | undefined;

  constructor(
    private http: HttpClient,
    private renderer: Renderer2,
    private router: Router,
    @Inject(DOCUMENT) private document: Document | undefined
  ) { }

  ngOnInit() {
    this.routerSubscription = this.router.events.subscribe(event => {
      if (event instanceof NavigationEnd) {
        this.updateBackground(event.urlAfterRedirects);
      }
    });
  }

  ngOnDestroy() {
    this.routerSubscription?.unsubscribe();
  }

  private updateBackground(url: string) {
    const containerElements = this.document?.querySelectorAll('.container');
    if (containerElements) {
      Array.from(containerElements).forEach((element: Element) => {
        if (url === '/login' || url === '/register') {
          this.renderer.removeClass(element, 'container_style');
        } else {
          this.renderer.addClass(element, 'container_style');
        }
      });
    }

    this.http.get('/api/login')
        .subscribe({
          next: () => {

          },
          error: (error) => {
            console.error('Error:', error);
          },
          complete: () => {
            // Optional: Handle completion logic if needed
          }
        });

  }
}
