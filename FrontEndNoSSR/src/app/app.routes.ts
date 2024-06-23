import { Routes } from '@angular/router';
import {HomeComponent} from "./pages/home/home.component";
import {ErrorComponent} from "./util/error/error.component";
import {LoginComponent} from "./pages/login/login.component";
import {DashboardComponent} from "./pages/dashboard/dashboard.component";
import {RegisterComponent} from "./pages/register/register.component";
import {authGuard} from "./auth.guard";

export const routes: Routes = [
  {path: "", component: HomeComponent},
  {path: "login", component: LoginComponent},
  {path: "register", component: RegisterComponent, },
  {path: "dashboard", component: DashboardComponent, canActivate: [authGuard]},
  {path: "**", component: ErrorComponent}
];
