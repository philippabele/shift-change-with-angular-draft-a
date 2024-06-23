import { CanActivateFn } from '@angular/router';
import { inject } from '@angular/core';
import { AuthStatusService } from './auth-status.service';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';
import { tap, map, first } from 'rxjs/operators';
import {LoginService} from "./login.service";

export const authGuard: CanActivateFn = (route, state) => {
    const authService = inject(LoginService);
    const authStatusService = inject(AuthStatusService);
    const router = inject(Router);

    console.log('Auth Guard triggered for route:', state.url); // Debugging log

    return authService.isLoggedIn().pipe(
        first(),
        tap(loggedIn => {
            console.log('Auth status:', loggedIn); // Debugging log
            authStatusService.updateAuthStatus(loggedIn);
            if (!loggedIn) {
                router.navigate(['/login']);
            }
        }),
        map(loggedIn => loggedIn) // Ensure the guard returns the final value
    ) as Observable<boolean>;
};
