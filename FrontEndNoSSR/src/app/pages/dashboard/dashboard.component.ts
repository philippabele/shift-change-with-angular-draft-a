import { Component } from '@angular/core';
import { HttpClient } from "@angular/common/http";
import { shift } from "../../shift.overview";
import { DatePipe, NgClass, NgForOf, NgIf } from "@angular/common";
import { FaqElementComponent } from "../../util/faq-element/faq-element.component";
import { FormsModule, NgForm } from "@angular/forms";

@Component({
    selector: 'app-dashboard',
    standalone: true,
    imports: [
        NgForOf,
        NgIf,
        FaqElementComponent,
        FormsModule,
        NgClass
    ],
    templateUrl: './dashboard.component.html',
    styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent {
    public shifts: shift[] = [];
    public newShift: Partial<shift> = {};  // Initialize newShift as an empty object
    public isOpen = false;
    public options = ['früh', 'spät', 'nachts'];
    public selectedOption: string | null = null;

    constructor(private http: HttpClient, private datePipe: DatePipe) {
        this.fetchShifts();
    }

    fetchShifts() {
        this.http.get<shift[]>('/api/shifts')
            .subscribe({
                next: (response) => {
                    this.shifts = response;
                },
                error: (error) => {
                    console.error('Error:', error);
                },
                complete: () => {
                    // Optional: Handle completion logic if needed
                }
            });
    }

    getFormattedDate(date: Date): string {
        return <string>this.datePipe.transform(date, 'dd.MM.yyyy');
    }

    save(shift: shift) {
        this.http.patch<shift>('/api/shifts/' + shift.uid, shift)
            .subscribe({
                next: (response) => {
                    this.fetchShifts();
                },
                error: (error) => {
                    console.error('Error:', error);
                    this.fetchShifts();
                },
                complete: () => {
                    // Optional: Handle completion logic if needed
                }
            });
    }

    delete(shift: shift) {
        this.http.delete('/api/shifts/' + shift.uid)
            .subscribe({
                next: () => {
                    this.fetchShifts();
                },
                error: (error) => {
                    console.error('Error:', error);
                    this.fetchShifts();
                },
                complete: () => {
                    // Optional: Handle completion logic if needed
                }
            });
    }

    toggleDropdown() {
        this.isOpen = !this.isOpen;
    }

    selectOption(option: string) {
        this.selectedOption = option;
        this.newShift.time = option;  // Set the selected option in newShift
        this.isOpen = false;
    }

    createShift(form: NgForm) {
        // Ensure that newShift has all necessary fields
        if (form.valid) {
            this.http.post('/api/shifts', this.newShift, { withCredentials: true })
                .subscribe({
                    next: (response) => {
                        this.fetchShifts();
                        this.newShift = {};  // Reset newShift after saving
                        this.selectedOption = null;  // Reset selected option
                        form.resetForm();  // Reset the form
                    },
                    error: (error) => {
                        console.error('Error:', error);
                    },
                    complete: () => {
                        // Optional: Handle completion logic if needed
                    }
                });
        } else {
            console.error('All fields are required.');
        }
    }
}
