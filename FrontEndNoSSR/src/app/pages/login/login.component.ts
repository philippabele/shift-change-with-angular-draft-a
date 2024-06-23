import { Component } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import {Router, RouterLink} from '@angular/router';
import {NgClass} from "@angular/common";

@Component({
  selector: 'app-login',
  standalone: true,
    imports: [
        FormsModule,
        NgClass,
        RouterLink
    ],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {
  constructor(private http: HttpClient, private router: Router) { }

  onSubmit(form: NgForm) {
    if (form.valid) {
      const inputData = { username: form.value.username, password: form.value.password };
      this.http.post('/api/login', inputData)
        .subscribe({
          next: () => {
            this.router.navigate(['/dashboard']);
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
}
