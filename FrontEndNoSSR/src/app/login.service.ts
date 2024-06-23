import { Injectable } from '@angular/core';
import {HttpClient} from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class LoginService {
  private apiUrl = '/api/login';

  constructor(private http: HttpClient) {
  }

  isLoggedIn(): Observable<boolean> {

    return new Observable<boolean>(observer => {
      this.http.get<{ loggedIn: boolean }>(this.apiUrl, { withCredentials: true }).subscribe({
        next: (response) => {
          console.log("Response: " + response.loggedIn)
          observer.next(response.loggedIn);
          observer.complete();
        },
        error: (error) => {
          observer.error(error);
        }
      });
    });
  }
}
