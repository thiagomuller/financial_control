import { Component, Input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CommonModule } from '@angular/common';
import { BankAccountSummary } from '../../../core/models/models';
import { TagChartComponent } from '../tag-chart/tag-chart.component';

@Component({
  selector: 'app-bank-account-card',
  standalone: true,
  imports: [CommonModule, RouterLink, TagChartComponent],
  templateUrl: './bank-account-card.component.html',
  styleUrl: './bank-account-card.component.css'
})
export class BankAccountCardComponent {
  @Input() summary!: BankAccountSummary;
}
