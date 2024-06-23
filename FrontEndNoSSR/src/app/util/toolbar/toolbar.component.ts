import { Component, OnDestroy, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';
import { NgIf } from '@angular/common';
import {LoginService} from "../../login.service";
import {AuthStatusService} from "../../auth-status.service";

@Component({
    selector: 'app-toolbar',
    standalone: true,
    imports: [
        RouterLink,
        NgIf
    ],
    templateUrl: './toolbar.component.html',
    styleUrl: './toolbar.component.scss'
})
export class ToolbarComponent implements OnInit, OnDestroy {
    private routerSubscription: Subscription | undefined;
    loggedIn = false;

    constructor(
        private loginService: LoginService,
        private router: Router,
        private authStatusService: AuthStatusService
    ) {
    }

    ngOnDestroy() {
        this.routerSubscription?.unsubscribe();
    }

    ngOnInit() {
        this.authStatusService.authStatus$.subscribe(isLoggedIn => {
            this.loggedIn = isLoggedIn;
            // Update component values based on auth status
        });
    }
}
