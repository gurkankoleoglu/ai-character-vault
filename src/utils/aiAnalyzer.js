// Basit AI analizi için kelimeler ve kategoriler
const categoryKeywords = {
  iş: ['toplanti', 'proje', 'deadline', 'sunum', 'rapor', 'müşteri', 'iş', 'ofis', 'yönetici', 'takım'],
  kişisel: ['aile', 'arkadaş', 'doğum günü', 'parti', 'yemek', 'sinema', 'spor', 'egzersiz', 'sağlık'],
  alışveriş: ['alışveriş', 'market', 'ürün', 'fiyat', 'indirim', 'sipariş', 'satın al', 'kredi kartı'],
  eğitim: ['ders', 'ödev', 'sınav', 'kitap', 'kurs', 'öğren', 'çalış', 'not', 'üniversite'],
  seyahat: ['seyahat', 'tatil', 'uçak', 'otobüs', 'otel', 'rezervasyon', 'bilet', 'yolculuk'],
};

const priorityKeywords = {
  yüksek: ['acil', 'önemli', 'hemen', 'bugün', 'kritik', 'deadline', 'son teslim'],
  orta: ['bu hafta', 'yakında', 'planla', 'hazırla'],
  düşük: ['ilgilen', 'belki', 'sonra', 'boş zaman'],
};

export const analyzeNote = (text) => {
  const lowerText = text.toLowerCase();
  
  // Kategori belirleme
  let category = 'diğer';
  let maxMatches = 0;
  
  for (const [cat, keywords] of Object.entries(categoryKeywords)) {
    const matches = keywords.filter(keyword => lowerText.includes(keyword)).length;
    if (matches > maxMatches) {
      maxMatches = matches;
      category = cat;
    }
  }
  
  // Öncelik belirleme
  let priority = 'orta';
  for (const [prio, keywords] of Object.entries(priorityKeywords)) {
    if (keywords.some(keyword => lowerText.includes(keyword))) {
      priority = prio;
      break;
    }
  }
  
  // Özetleme (basit - ilk 50 karakter)
  const summary = text.length > 50 ? text.substring(0, 50) + '...' : text;
  
  return {
    category,
    priority,
    summary,
  };
};

export const getCategoryColor = (category) => {
  const colors = {
    iş: 'bg-blue-100 text-blue-800',
    kişisel: 'bg-green-100 text-green-800',
    alışveriş: 'bg-yellow-100 text-yellow-800',
    eğitim: 'bg-purple-100 text-purple-800',
    seyahat: 'bg-pink-100 text-pink-800',
    diğer: 'bg-gray-100 text-gray-800',
  };
  return colors[category] || colors.diğer;
};

export const getPriorityColor = (priority) => {
  const colors = {
    yüksek: 'bg-red-100 text-red-800',
    orta: 'bg-orange-100 text-orange-800',
    düşük: 'bg-green-100 text-green-800',
  };
  return colors[priority] || colors.orta;
};
