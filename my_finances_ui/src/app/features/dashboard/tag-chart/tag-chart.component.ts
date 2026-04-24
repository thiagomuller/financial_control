import { Component, Input, OnChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TagStat } from '../../../core/models/models';

interface Segment { color: string; dasharray: string; dashoffset: string; label: string; count: number; }

const FALLBACK_COLORS = ['#2563eb','#16a34a','#dc2626','#d97706','#7c3aed','#0891b2','#be185d','#65a30d'];

@Component({
  selector: 'app-tag-chart',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './tag-chart.component.html',
  styleUrl: './tag-chart.component.css'
})
export class TagChartComponent implements OnChanges {
  @Input() tagStats: TagStat[] = [];

  segments: Segment[] = [];
  total = 0;

  ngOnChanges(): void {
    this.total = this.tagStats.reduce((s, ts) => s + ts.count, 0);
    if (this.total === 0) { this.segments = []; return; }

    const circumference = 100;
    let offset = 25; // start at top (offset 25 = -90deg rotation trick)

    this.segments = this.tagStats.map((ts, i) => {
      const frac = ts.count / this.total;
      const dash = frac * circumference;
      const seg: Segment = {
        color: ts.tag.color ?? FALLBACK_COLORS[i % FALLBACK_COLORS.length],
        dasharray: `${dash} ${circumference - dash}`,
        dashoffset: `${offset}`,
        label: ts.tag.name,
        count: ts.count
      };
      offset -= dash;
      return seg;
    });
  }
}
