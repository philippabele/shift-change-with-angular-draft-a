import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FaqElementComponent } from './faq-element.component';

describe('FaqElementComponent', () => {
  let component: FaqElementComponent;
  let fixture: ComponentFixture<FaqElementComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FaqElementComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(FaqElementComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
